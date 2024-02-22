import { Shop } from '@/assets/icons/circle';
import { Meta, StoryObj } from '@storybook/react';
import { useState } from 'react';
import { Header } from './Header';
import { LabelledIcon } from './LabelledIcon';
import { Line } from './Line';
import { Stack as StackComponent } from './Stack';

type Story = StoryObj<typeof StackComponent>;

const meta: Meta<typeof StackComponent> = {
    title: 'Primitives/Stack',
};

const StackStory = ({ collapsible }: { collapsible?: boolean }) => {
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
                <Line>
                    <LabelledIcon>
                        <LabelledIcon.Icon>
                            <Shop />
                        </LabelledIcon.Icon>
                        <LabelledIcon.Label>Collection</LabelledIcon.Label>
                    </LabelledIcon>
                </Line>
                <Line>
                    <LabelledIcon>
                        <LabelledIcon.Icon>
                            <Shop />
                        </LabelledIcon.Icon>
                        <LabelledIcon.Label>Collection</LabelledIcon.Label>
                    </LabelledIcon>
                </Line>
                <Line>
                    <LabelledIcon>
                        <LabelledIcon.Icon>
                            <Shop />
                        </LabelledIcon.Icon>
                        <LabelledIcon.Label>Collection</LabelledIcon.Label>
                    </LabelledIcon>
                </Line>
            </StackComponent.Content>
        </StackComponent>
    );
};

export const Default: Story = {
    render: () => <StackStory />,
};

export const Collapsible: Story = {
    render: () => <StackStory collapsible />,
};

export default meta;
