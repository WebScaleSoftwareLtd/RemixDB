// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

type Props = {
    error: Error;
};

export default ({
    error,
}: Props) => <div className="flex flex-col items-center justify-center">
    <h1 className="text-3xl">
        <XIcon className="inline-block w-8 h-8" />
    </h1>
    <h2 className="text-xl">Something went wrong</h2>
    <p className="text-gray-500">{error.message}</p>
</div>;

function XIcon(props: React.SVGProps<SVGSVGElement>) {
    return <svg
        {...props}
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
    >
        <path d="M18 6 6 18" />
        <path d="m6 6 12 12" />
    </svg>;
}
