import resolve from "@rollup/plugin-node-resolve";
import cjs from "@rollup/plugin-commonjs";

export default {
    input: "main.js",
    output: {
        file: "bundle.js",
        format: "iife",
    },
	plugins: [
        resolve(),
        cjs()
    ]
};
