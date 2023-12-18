module.exports = {
    extends: [
        "semistandard",
        "plugin:@typescript-eslint/recommended",
    ],
    rules: {
        semi: ["error", "always"],
        quotes: ["error", "double"],
        indent: ["error", 4],
        "no-extra-parens": "error",
        "@typescript-eslint/no-explicit-any": "off",
        "func-call-spacing": "off",
    },
    plugins: ["@typescript-eslint"],
};
