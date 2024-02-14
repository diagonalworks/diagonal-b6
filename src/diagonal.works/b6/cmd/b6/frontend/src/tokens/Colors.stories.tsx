import { ColorItem, ColorPalette } from '@storybook/blocks';
import type { Meta } from '@storybook/react';
import { hsl } from 'd3-color';
import resolveConfig from 'tailwindcss/resolveConfig';

import { toTitleCase } from '@/lib/text';
import { $FixMe } from '@/utils/defs';
import tailwindConfig from '../../tailwind.config';
const fullConfig = resolveConfig(tailwindConfig);

const colorOrder = Object.entries(fullConfig.theme.colors)
    .map(([name, value]: [string, $FixMe]) => {
        return {
            name,
            value: Object.values(value)[0] as string,
        };
    })
    .sort((a, b) => hsl(a.value).h - hsl(b.value).h)
    .map((c) => c.name);

const sortedColors = Object.entries(fullConfig.theme.colors).sort(
    ([aName], [bName]) => colorOrder.indexOf(aName) - colorOrder.indexOf(bName),
);

export const Colors = () => {
    return (
        <ColorPalette>
            {sortedColors.map(([name, value]) => {
                return (
                    <ColorItem
                        key={name}
                        subtitle=""
                        title={toTitleCase(name)}
                        colors={value}
                    />
                );
            })}
        </ColorPalette>
    );
};

const meta = {
    title: 'Tokens/Colors',
} as Meta;

export default meta;
