import { Shop } from '@/assets/icons/circle';
import type { Meta, StoryObj } from '@storybook/react';
import { useState } from 'react';
import { LabelledIcon as LabelledIconComponent } from './LabelledIcon';
import { Line as LineComponent } from './Line';
import { Select as SelectComponent } from './Select';

type Story = StoryObj<typeof LineComponent>;

export const LabelledIcon: Story = {
    render: () => (
        <div className=" border border-graphite-30 border-dashed w-fit p-2">
            <LabelledIconComponent slots={{ icon: <Shop /> }}>
                Hello
            </LabelledIconComponent>
        </div>
    ),
};

const OPTIONS = ['5', '10', '15'];

const SelectStory = () => {
    const [value, setValue] = useState(OPTIONS[0]);

    const label = (value: string) => {
        return `${value} min`;
    };

    return (
        <div className=" border border-graphite-30 border-dashed w-fit p-2">
            <SelectComponent value={value} onValueChange={setValue}>
                <SelectComponent.Button>
                    <SelectComponent.Primitive.Value>
                        {label(value)}
                    </SelectComponent.Primitive.Value>
                </SelectComponent.Button>
                <SelectComponent.Options>
                    {OPTIONS.map((option) => (
                        <SelectComponent.Option key={option} value={option}>
                            {label(option)}
                        </SelectComponent.Option>
                    ))}
                </SelectComponent.Options>
            </SelectComponent>
        </div>
    );
};

export const Select: Story = {
    render: () => <SelectStory />,
};

const meta: Meta = {
    title: 'Atoms/Atom',
};

export default meta;
