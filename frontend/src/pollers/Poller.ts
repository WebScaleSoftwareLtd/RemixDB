// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { authenticated, subscribeAuthenticated } from "@/authState";

export default class Poller<T> {
    private timeoutHandler: any;
    private _events: { at: number; event: T }[] = [];
    private _err: Error | undefined;

    get events() {
        if (this._err) throw this._err;
        return this._events;
    }

    private subscribers: Map<number, () => void> = new Map();
    private subscriberId = 0;

    subscribe(cb: () => void): number {
        const id = this.subscriberId++;
        this.subscribers.set(id, cb);
        return id;
    }

    unsubscribe(id: number) {
        this.subscribers.delete(id);
    }

    private emitEvent() {
        for (const cb of this.subscribers.values()) cb();
    }

    constructor(private fetcher: () => Promise<T>, private tickDelay: number) {
        this.handleAuthState(authenticated);
        subscribeAuthenticated(this.handleAuthState.bind(this));
    }

    private handleAuthState(authenticated: boolean) {
        // If we aren't authenticated, kill the interval handler.
        if (!authenticated) {
            clearTimeout(this.timeoutHandler);
            this.timeoutHandler = undefined;
            this._events = [];
            this._err = undefined;
            return;
        }

        // Invoke a tick here.
        this.tick();
    }

    private async tick() {
        try {
            // Try to do the event.
            this.events.push({
                at: Date.now(),
                event: await this.fetcher(),
            });
        } catch (e) {
            // If the request fails, set the error and return.
            this._err = e as Error;
            return;
        } finally {
            // In either case, emit a event here.
            this.emitEvent();
        }

        // Set a timeout to invoke another tick.
        this.timeoutHandler = setTimeout(this.tick.bind(this), this.tickDelay);
    }
}
