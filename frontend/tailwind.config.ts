import type { Config } from 'tailwindcss';
import defaultTheme from 'tailwindcss/defaultTheme';
import colors from './src/tokens/colors.json';

const defaultFontFamily = ['"Unica 77"', ...defaultTheme.fontFamily.sans];

const config: Config = {
    content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
    theme: {
        ...defaultTheme,
        fontFamily: {
            sans: defaultFontFamily,
        },
        colors: {
            ...colors,
            white: 'white',
            transparent: 'transparent',
        },
        extend: {
            animation: {
                blink: 'blink 1s infinite',
            },
            keyframes: {
                blink: {
                    from: { 'border-left-color': 'transparent' },
                    to: { 'border-left-color': colors.ultramarine[70] },
                },
            },
        },
    },
    plugins: [],
};

export default config;
