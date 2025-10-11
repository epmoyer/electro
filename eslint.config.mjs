import globals from "globals";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const compat = new FlatCompat({
    baseDirectory: __dirname,
    recommendedConfig: js.configs.recommended,
    allConfig: js.configs.all
});

export default [...compat.extends("eslint:recommended"), {
    languageOptions: {
        globals: {
            ...globals.browser,
            $: true,
        },

        ecmaVersion: 12,
        sourceType: "script",
    },

    rules: {
        indent: ["error", 4],
        // "react/jsx-indent": ["error", 4],
        // "react/jsx-indent-props": ["error", 4],
        semi: ["error", "always"],
        "comma-dangle": ["error", "only-multiline"],
    },
}];