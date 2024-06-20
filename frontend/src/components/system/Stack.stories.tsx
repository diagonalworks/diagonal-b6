import { Meta, StoryObj } from '@storybook/react';
import { useState } from 'react';

import { Shop } from '@/assets/icons/circle';
import { Header } from '@/components/system/Header';
import { LabelledIcon } from '@/components/system/LabelledIcon';
import { Line } from '@/components/system/Line';
import { Stack as StackComponent } from '@/components/system/Stack';

type Story = StoryObj<typeof StackComponent>;

const meta: Meta<typeof StackComponent> = {
    title: 'Primitives/Stack',
};

const StackStory = ({
    collapsible,
    collections = 0,
}: {
    collapsible?: boolean;
    collections?: number;
}) => {
    const [open, setOpen] = useState(false);

    return (
        <StackComponent
            open={open}
            onOpenChange={setOpen}
            collapsible={collapsible}
        >
            <StackComponent.Trigger asChild>
                <Line>
                    <Header>
                        <Header.Label>Header</Header.Label>
                        <Header.Actions close share />
                    </Header>
                </Line>
            </StackComponent.Trigger>
            <StackComponent.Content>
                {Array.from({ length: collections }).map((_, i) => (
                    <Line key={i}>
                        <LabelledIcon>
                            <LabelledIcon.Icon>
                                <Shop />
                            </LabelledIcon.Icon>
                            <LabelledIcon.Label>Collection</LabelledIcon.Label>
                        </LabelledIcon>
                    </Line>
                ))}
            </StackComponent.Content>
        </StackComponent>
    );
};

export const Default: Story = {
    render: () => <StackStory collections={3} />,
};

export const Collapsible: Story = {
    render: () => <StackStory collapsible collections={3} />,
};

export const Scrollable: Story = {
    render: () => <StackStory collapsible collections={8} />,
};

export default meta;
