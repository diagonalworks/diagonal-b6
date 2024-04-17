import { appAtom } from '@/atoms/app';
import { Header } from '@/components/system/Header';
import { LabelledIcon } from '@/components/system/LabelledIcon';
import { Line } from '@/components/system/Line';
import { Select } from '@/components/system/Select';
import { Stack } from '@/components/system/Stack';
import {
    AtomProto,
    ChoiceLineProto,
    LabelledIconProto,
    LineProto,
    StackProto,
    SubstackProto,
} from '@/types/generated/ui';
import { DotIcon, FrameIcon, SquareIcon } from '@radix-ui/react-icons';
import { useSetAtom } from 'jotai';
import { omit } from 'lodash';
import { createContext, useContext, useMemo, useState } from 'react';
import { match } from 'ts-pattern';

const StackContext = createContext<{
    id?: string;
}>({});

const useStackContext = () => useContext(StackContext);

export const StackWrapper = ({
    stack,
    id,
}: {
    stack: StackProto;
    id: string;
}) => {
    return (
        <StackContext.Provider
            value={{
                id,
            }}
        >
            {stack.substacks.map((substack, i) => {
                return <SubstackWrapper key={i} substack={substack} />;
            })}
        </StackContext.Provider>
    );
};

export const SubstackWrapper = ({ substack }: { substack: SubstackProto }) => {
    return (
        <Stack collapsible={substack.collapsable}>
            {substack.lines.map((l, i) => {
                return <LineWrapper key={i} line={l} />;
            })}
        </Stack>
    );
};

export const LineWrapper = ({ line }: { line: LineProto }) => {
    const setAppAtom = useSetAtom(appAtom);
    const { id } = useStackContext();
    return (
        <Line>
            {line?.header && (
                <Header>
                    {line.header.title && (
                        <Header.Label>
                            <AtomWrapper atom={line.header.title} />
                        </Header.Label>
                    )}
                    <Header.Actions
                        close
                        slotProps={{
                            close: {
                                onClick: () => {
                                    if (!id) return;
                                    setAppAtom((draft) => {
                                        draft.stacks = omit(draft.stacks, id);
                                    });
                                },
                            },
                        }}
                    />
                </Header>
            )}
            {line.choice && <SelectWrapper choice={line.choice} />}
            {line.value && line.value.atom && (
                <AtomWrapper atom={line.value.atom} />
            )}
        </Line>
    );
};

export const AtomWrapper = ({ atom }: { atom: AtomProto }) => {
    return (
        <>
            {atom.labelledIcon && (
                <LabelledIconWrapper labelledIcon={atom.labelledIcon} />
            )}
            {atom.value && <Line.Value>{atom.value}</Line.Value>}
            {/* @TODO: render other primitive atom types */}
        </>
    );
};

export const SelectWrapper = ({ choice }: { choice: ChoiceLineProto }) => {
    const options = useMemo(
        () =>
            choice.chips.map((chip) => ({
                value: chip.value,
                label: chip.labelledIcon?.label ?? chip.value,
            })),
        [choice.chips]
    );
    /* This component will need to be controlled, but unsure how exactly yet. */
    const [value, setValue] = useState(options[0].value);

    const label = (value: string) => {
        return options.find((option) => option.value === value)?.label ?? '';
    };

    return (
        <Select value={value} onValueChange={setValue}>
            <Select.Button>{value && label(value)}</Select.Button>
            <Select.Options>
                {options.map((option) => (
                    <>
                        {option.value && (
                            <Select.Option
                                key={option.value}
                                value={option.value}
                            >
                                {option.label}
                            </Select.Option>
                        )}
                    </>
                ))}
            </Select.Options>
        </Select>
    );
};

const LabelledIconWrapper = ({
    labelledIcon,
}: {
    labelledIcon: LabelledIconProto;
}) => {
    const icon = match(labelledIcon.icon)
        .with('area', () => <FrameIcon />)
        .with('point', () => <DotIcon />)
        .otherwise(() => <SquareIcon />);

    return (
        <LabelledIcon>
            <LabelledIcon.Icon className=" text-ultramarine-60">
                {icon}
            </LabelledIcon.Icon>
            <LabelledIcon.Label>{labelledIcon.label}</LabelledIcon.Label>
        </LabelledIcon>
    );
};
