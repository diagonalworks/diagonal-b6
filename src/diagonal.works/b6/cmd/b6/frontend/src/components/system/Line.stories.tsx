import { Grocery, School, Shop } from '@/assets/icons/circle';
import type { Meta, StoryObj } from '@storybook/react';
import { useState } from 'react';
import { Header } from './Header';
import { LabelledIcon } from './LabelledIcon';
import { Line as LineComponent } from './Line';
import { Select } from './Select';

type Story = StoryObj<typeof LineComponent>;

const SELECT_OPTIONS = {
    travel: [
        { value: '15-walk', label: '15 min walk' },
        { value: '30-walk', label: '30 min walk' },
        { value: '20-bus', label: '20 min bus' },
    ],

    grocery: [
        { value: 'all', label: 'all' },
        { value: 'convenience', label: 'convenience shops' },
        { value: 'comparison', label: 'comparison shops' },
    ],
};

const SelectForStory = ({
    type,
}: {
    type: 'travel' | 'grocery';
    className?: string;
}) => {
    const options = SELECT_OPTIONS[type];
    const [value, setValue] = useState(options[0].value);

    const label = (value: string) => {
        return options.find((option) => option.value === value)?.label ?? '';
    };

    return (
        <Select value={value} onValueChange={setValue}>
            <Select.Button>{label(value)}</Select.Button>
            <Select.Options>
                {options.map((option) => (
                    <Select.Option key={option.value} value={option.value}>
                        {option.label}
                    </Select.Option>
                ))}
            </Select.Options>
        </Select>
    );
};

export const Line: Story = {
    render: () => (
        <div className="flex flex-col gap-8">
            <div>
                <h3 className="mb-2">Empty Line</h3>
                <LineComponent>
                    <div className="text-sm text-graphite-40">
                        {'< line contents >'}
                    </div>
                </LineComponent>
            </div>
            <div>
                <h3 className="mb-2">Lines with Atoms</h3>
                <div className="flex flex-col gap-2">
                    <LineComponent>
                        <LabelledIcon>
                            <LabelledIcon.Icon>
                                <Shop />
                            </LabelledIcon.Icon>
                            <LabelledIcon.Label>Collection</LabelledIcon.Label>
                        </LabelledIcon>
                    </LineComponent>
                    <LineComponent className="justify-between">
                        <LabelledIcon>
                            <LabelledIcon.Icon>
                                <School />
                            </LabelledIcon.Icon>
                            <LabelledIcon.Label>Schools</LabelledIcon.Label>
                        </LabelledIcon>
                        <LineComponent.Value>3</LineComponent.Value>
                    </LineComponent>
                    <LineComponent>
                        <LabelledIcon>
                            <LabelledIcon.Icon>
                                <Grocery />
                            </LabelledIcon.Icon>
                            <LabelledIcon.Label>Collection</LabelledIcon.Label>
                        </LabelledIcon>
                        <div className="flex items-center min-w-0 gap-1">
                            <SelectForStory type="travel" />
                            <SelectForStory type="grocery" />
                        </div>
                    </LineComponent>
                    <LineComponent>
                        <LabelledIcon>
                            <LabelledIcon.Icon>
                                <Grocery />
                            </LabelledIcon.Icon>
                            <LabelledIcon.Label>
                                a very long collection name
                            </LabelledIcon.Label>
                        </LabelledIcon>
                        <div className="flex items-center min-w-0 gap-1 ">
                            <SelectForStory type="travel" />
                            <SelectForStory type="grocery" />
                        </div>
                    </LineComponent>
                    <LineComponent>
                        <Header>
                            <Header.Label>Header</Header.Label>
                            <Header.Actions close share />
                        </Header>
                    </LineComponent>
                </div>
            </div>
        </div>
    ),
};

const meta: Meta<typeof LineComponent> = {
    component: LineComponent,
    title: 'Primitives/Line',
};

export default meta;
