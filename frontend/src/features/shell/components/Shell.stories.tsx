import type { Meta, StoryObj } from '@storybook/react';

import { Shell as ShellComponent } from '@/features/shell/components/Shell';

type Story = StoryObj<typeof ShellComponent>;

export const Shell: Story = {
    render: () => {
        return (
            <ShellComponent
                functions={[
                    { id: 'add-tag', description: 'Add a tag to a resource' },
                    {
                        id: 'centroid',
                        description: 'Compute the centroid of a geometry',
                    },
                    {
                        id: 'closest',
                        description: 'Find the closest resource to a point',
                    },
                ]}
            />
        );
    },
};

const meta: Meta = {
    title: 'Components/Shell',
};

export default meta;
