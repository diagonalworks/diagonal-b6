import type { Config } from 'tailwindcss';
import defaultTheme from 'tailwindcss/defaultTheme';
import colors from './src/tokens/colors.json';

const defaultFontFamily = ['"Unica 77"', ...defaultTheme.fontFamily.sans];

const config: Config = {
    content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
    theme: {
        fontFamily: {
            sans: defaultFontFamily,
        },
        colors: {
            ...colors,
        },
        extend: {},
    },
    plugins: [],
};

export default config;
