import { color } from 'd3-color';
import { Color } from 'deck.gl/typed';

export const rgbToHex = (r: number, g: number, b: number) =>
    '#' +
    [r, g, b]
        .map((x) => {
            const hex = x.toString(16);
            return hex.length === 1 ? '0' + hex : hex;
        })
        .join('');

export const colorToRgbArray = (c: string, alpha: number = 255) => {
    const rgbColor = color(c)?.rgb();
    if (!rgbColor) return [0, 0, 0, 0] as Color;
    return [rgbColor.r, rgbColor.g, rgbColor.b, alpha] as Color;
};
