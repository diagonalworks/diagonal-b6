import * as circleIcons from '@/assets/icons/circle';
import { appAtom } from '@/atoms/app';
import { Header } from '@/components/system/Header';
import { LabelledIcon } from '@/components/system/LabelledIcon';
import { Line } from '@/components/system/Line';
import { Select } from '@/components/system/Select';
import { Stack } from '@/components/system/Stack';
import { Tooltip } from '@/components/system/Tooltip';
import {
    AtomProto,
    ChipProto,
    ConditionalProto,
    LabelledIconProto,
    LineProto,
    StackProto,
    SubstackProto,
} from '@/types/generated/ui';
import { $FixMe } from '@/utils/defs';
import {
    DotIcon,
    ExclamationTriangleIcon,
    FrameIcon,
    SquareIcon,
} from '@radix-ui/react-icons';
import { useSetAtom } from 'jotai';
import { isObject, isUndefined, omit } from 'lodash';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useMemo,
} from 'react';
import { match } from 'ts-pattern';
import { Updater, useImmer } from 'use-immer';

const StackContext = createContext<{
    id?: string;
}>({});

const useStackContext = () => useContext(StackContext);

const LineContext = createContext<{
    state: LineStore;
    setState: Updater<LineStore>;
    setChipValue: (index: number, value: number) => void;
}>({
    state: { line: {}, chips: {} },
    setState: () => {},
    setChipValue: () => {},
});

const useLineContext = () => useContext(LineContext);

/**
 * Recursively find atoms in a line. If a type is provided, only atoms of that type will be returned.
 * This function is currently a mess because types of line elements are loosely defined.
 */
const findAtoms = (line: $FixMe, type?: keyof AtomProto): AtomProto[] => {
    const atom = line?.atom;
    if (atom) {
        if (type) {
            return atom?.[type] ? [atom] : [];
        }
        return [atom];
    }

    if (Array.isArray(line)) {
        return line.flatMap((l) => findAtoms(l, type));
    }

    if (isObject(line)) {
        return Object.keys(line).flatMap((key) =>
            findAtoms((line as $FixMe)[key], type)
        );
    }
    return [];
};

export type Chip = { atom: ChipProto; value: number };

export type LineStore = {
    line: LineProto;
    chips: Record<number, Chip>;
};

const LineContextProvider = ({
    line,
    children,
}: {
    line: LineProto;
} & PropsWithChildren) => {
    const chips = useMemo(() => {
        const chipMap: LineStore['chips'] = {};

        findAtoms(line, 'chip').forEach((atom) => {
            if (atom.chip) {
                if (isUndefined(atom.chip.index)) {
                    console.warn(`Chip index is undefined`, { line, atom });
                }
                chipMap[atom.chip.index] = {
                    atom: atom.chip,
                    value: 0,
                };
            }
        });

        return chipMap;
    }, [line]);

    const [state, setState] = useImmer<LineStore>({
        line,
        chips,
    });

    const setChipValue = useCallback(
        (index: number, value: number) => {
            setState((draft) => {
                if (!draft.chips[index]) return;
                draft.chips[index].value = value;
            });
        },
        [setState]
    );

    const lineContextStoreData = useMemo(() => {
        return {
            state,
            setState,
            setChipValue,
        };
    }, [state, setState, setChipValue]);
    return (
        <LineContext.Provider value={lineContextStoreData}>
            {children}
        </LineContext.Provider>
    );
};

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
        <LineContextProvider line={line}>
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
                                            draft.stacks = omit(
                                                draft.stacks,
                                                id
                                            );
                                        });
                                    },
                                },
                            }}
                        />
                    </Header>
                )}
                {/* {line.choice && <SelectWrapper choice={line.choice} />} */}
                {line.value && line.value.atom && (
                    <AtomWrapper atom={line.value.atom} />
                )}
                {line.leftRightValue && (
                    <div className="justify-between flex items-center w-full">
                        <div className="flex items-center gap-2 w-11/12 flex-grow-0">
                            {line.leftRightValue.left.map(({ atom }, i) => {
                                if (!atom) return null;
                                return <AtomWrapper key={i} atom={atom} />;
                            })}
                        </div>
                        {line.leftRightValue.right?.atom && (
                            <div className="flex items-center gap-1">
                                <AtomWrapper
                                    atom={line.leftRightValue.right.atom}
                                />
                            </div>
                        )}
                    </div>
                )}
            </Line>
        </LineContextProvider>
    );
};

export const ChipWrapper = ({
    chip,
    onChange,
}: {
    chip: Chip;
    onChange: (v: number) => void;
}) => {
    const options = useMemo(
        () =>
            chip.atom.labels.map((label, i) => ({
                value: i.toString(),
                label,
            })),
        [chip.atom.labels]
    );

    return (
        <SelectWrapper
            options={options}
            value={chip.value.toString()}
            onValueChange={(v) => onChange(parseInt(v))}
        />
    );
};

export const AtomWrapper = ({ atom }: { atom: AtomProto }) => {
    const line = useLineContext();

    if (atom.value) {
        return <Line.Value>{atom.value}</Line.Value>;
    }

    if (atom.labelledIcon) {
        return <LabelledIconWrapper labelledIcon={atom.labelledIcon} />;
    }

    if (atom.chip) {
        const lineChip = line.state.chips[atom.chip.index];
        if (lineChip) {
            return (
                <ChipWrapper
                    chip={lineChip}
                    onChange={(value: number) =>
                        line.setChipValue(lineChip.atom.index, value)
                    }
                />
            );
        }
    }

    if (atom.conditional) {
        return <ConditionalWrapper atom={atom.conditional} />;
    }

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

const ConditionalWrapper = ({ atom }: { atom: ConditionalProto }) => {
    const line = useLineContext();

    const atomIndex = atom.conditions.findIndex((condition) => {
        const check = condition.indices.map((index, i) => {
            const chip = line.state.chips[index];
            if (!chip) return false;
            return chip.value === condition.values[i];
        });
        return check.every(Boolean);
    });

    if (atomIndex === -1)
        return (
            <Tooltip content="Value not found">
                <ExclamationTriangleIcon className="text-graphite-50" />
            </Tooltip>
        );

    return <AtomWrapper atom={atom.atoms[atomIndex]} />;
};

export const SelectWrapper = ({
    options,
    value,
    onValueChange,
}: {
    options: { value: string; label: string }[];
    value: string;
    onValueChange: (v: string) => void;
}) => {
    const label = (value: string) => {
        return options.find((option) => option.value === value)?.label ?? '';
    };

    return (
        <Select value={value} onValueChange={onValueChange}>
            <Select.Button>{value && label(value)}</Select.Button>
            <Select.Options>
                {options.map((option, i) => (
                    <div key={i}>
                        {option.value && (
                            <Select.Option value={option.value}>
                                {option.label}
                            </Select.Option>
                        )}
                    </div>
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
        .otherwise(() => {
            const iconComponentName = `${labelledIcon.icon
                .charAt(0)
                .toUpperCase()}${labelledIcon.icon.slice(1)}`;

            if (circleIcons[iconComponentName as keyof typeof circleIcons]) {
                const Icon =
                    circleIcons[iconComponentName as keyof typeof circleIcons];
                return <Icon />;
            }
            return <SquareIcon />;
        });

    return (
        <LabelledIcon>
            <LabelledIcon.Icon className=" text-ultramarine-60">
                {icon}
            </LabelledIcon.Icon>
            {/* otherwise hard for elements to fit in line */}
            <LabelledIcon.Label className="text-sm">
                {labelledIcon.label}
            </LabelledIcon.Label>
        </LabelledIcon>
    );
};
