import type { Config } from "tailwindcss";
import colors from "./src/tokens/colors.json";

const config: Config =  {
    content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
    theme: {
       colors: {
        ...colors
       },
       extend: {},
    },
    plugins: [],
};


export default config;
