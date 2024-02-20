import { Shop } from '@/assets/icons/circle';
import type { Meta, StoryObj } from '@storybook/react';
import { LabelledIcon } from './LabelledIcon';
import { Line as LineComponent } from './Line';

type Story = StoryObj<typeof LineComponent>;

export const Line: Story = {
    render: () => (
        <div className="flex flex-col gap-4">
            <LineComponent>
                <div className="text-sm text-graphite-40">
                    {'< line contents >'}
                </div>
            </LineComponent>
            <div>
                <h3 className="mb-1">Line with Atoms</h3>
                <LineComponent>
                    <LabelledIcon>
                        <LabelledIcon.Icon>
                            <Shop />
                        </LabelledIcon.Icon>
                        <LabelledIcon.Label>Collection</LabelledIcon.Label>
                    </LabelledIcon>
                </LineComponent>
            </div>
        </div>
    ),
};

const meta: Meta<typeof LineComponent> = {
    component: LineComponent,
    title: 'Primitives/Line',
};

export default meta;
