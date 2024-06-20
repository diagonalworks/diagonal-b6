import { ColorItem, ColorPalette } from '@storybook/blocks';
import type { Meta } from '@storybook/react';
import { hsl } from 'd3-color';

import { $FixMe, $IntentionalAny } from '@/utils/defs';
import { toTitleCase } from '@/utils/text';

import colors from './colors.json';

const colorOrder = Object.entries(colors)
    .map(([name, value]: [string, $FixMe]) => {
        return {
            name,
            value: Object.values(value)[0] as string,
        };
    })
    .sort((a, b) => hsl(a.value).h - hsl(b.value).h)
    .map((c) => c.name);

const sortedColors = Object.entries(colors).sort(
    ([aName], [bName]) => colorOrder.indexOf(aName) - colorOrder.indexOf(bName)
);

export const Colors = () => {
    return (
        <div className="flex flex-col gap-8">
            <h1 className="pb-1 text-sm border-b text-violet-70 border-violet-70 ">
                Colors
            </h1>

            <ColorPalette>
                {sortedColors.map(([name, value]) => {
                    return (
                        <ColorItem
                            key={name}
                            subtitle=""
                            title={toTitleCase(name)}
                            colors={value as $IntentionalAny}
                        />
                    );
                })}
            </ColorPalette>
        </div>
    );
};

const meta = {
    title: 'Tokens/Colors',
} as Meta;

export default meta;
