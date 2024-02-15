import type { Meta, StoryObj } from '@storybook/react';
import { Line as LineComponent } from './Line';

type Story = StoryObj<typeof LineComponent>;

export const Line: Story = {
    render: () => (
        <LineComponent>
            <div className="text-sm text-graphite-40">
                {'< line contents >'}
            </div>
        </LineComponent>
    ),
};

const meta: Meta<typeof LineComponent> = {
    component: LineComponent,
    title: 'Atoms/Line',
};

export default meta;
