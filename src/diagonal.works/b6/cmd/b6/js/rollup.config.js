import resolve from "@rollup/plugin-node-resolve";
import cjs from "@rollup/plugin-commonjs";

export default {
    input: "b6.js",
    output: {
        file: "bundle.js",
        format: "iife",
    },
	plugins: [
		resolve(),
        cjs()
    ]
};
