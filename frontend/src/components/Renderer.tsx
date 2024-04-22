import * as circleIcons from '@/assets/icons/circle';
import { AppStore, appAtom } from '@/atoms/app';
import { Header } from '@/components/system/Header';
import { LabelledIcon } from '@/components/system/LabelledIcon';
import { Line } from '@/components/system/Line';
import { Select } from '@/components/system/Select';
import { Stack } from '@/components/system/Stack';
import { Tooltip } from '@/components/system/Tooltip';
import { fetchB6 } from '@/lib/b6';
import { LineContextProvider, useLineContext } from '@/lib/context/line';
import { StackContextProvider, useStackContext } from '@/lib/context/stack';
import colors from '@/tokens/colors.json';
import {
    AtomProto,
    ChoiceProto,
    ConditionalProto,
    HeaderLineProto,
    LabelledIconProto,
    LineProto,
    SubstackProto,
    SwatchLineProto,
} from '@/types/generated/ui';
import { Chip, StackResponse } from '@/types/stack';
import { $FixMe } from '@/utils/defs';
import {
    DotIcon,
    ExclamationTriangleIcon,
    FrameIcon,
    SquareIcon,
} from '@radix-ui/react-icons';
import { useQuery } from '@tanstack/react-query';
import { scaleOrdinal } from '@visx/scale';
import { interpolateRgbBasis } from 'd3-interpolate';
import { useAtom, useSetAtom } from 'jotai';
import { omit } from 'lodash';
import React, { useEffect, useMemo, useState } from 'react';
import { useMap } from 'react-map-gl/maplibre';
import { match } from 'ts-pattern';
import { Histogram } from './system/Histogram';

export const StackWrapper = ({
    stack,
    docked = false,
    mapId,
}: {
    stack: AppStore['stacks'][string];
    id: string;
    docked?: boolean;
    mapId: string;
}) => {
    const [open, setOpen] = useState(docked ? false : true);
    if (!stack.proto.stack) return null;

    const firstSubstack = stack.proto.stack.substacks[0];
    const otherSubstacks = stack.proto.stack.substacks.slice(1);

    return (
        <div>
            <StackContextProvider stack={stack} mapId={mapId}>
                <Stack collapsible={docked} open={open} onOpenChange={setOpen}>
                    <Stack.Trigger>
                        <SubstackWrapper
                            substack={firstSubstack}
                            collapsible={firstSubstack.collapsable}
                        />
                    </Stack.Trigger>
                    <Stack.Content>
                        {otherSubstacks.map((substack, i) => {
                            return (
                                <SubstackWrapper
                                    key={i}
                                    substack={substack}
                                    collapsible={substack.collapsable}
                                />
                            );
                        })}
                    </Stack.Content>
                </Stack>
            </StackContextProvider>
        </div>
    );
};

export const SubstackWrapper = ({
    substack,
    collapsible = false,
}: {
    substack: SubstackProto;
    collapsible?: boolean;
}) => {
    const header = substack.lines[0]?.header;
    const contentLines = substack.lines.slice(header ? 1 : 0);
    const [open, setOpen] = useState(collapsible ? false : true);

    const isHistogram =
        contentLines.length > 1 && contentLines.every((line) => line.swatch);

    return (
        <Stack
            collapsible={substack.collapsable}
            open={open}
            onOpenChange={setOpen}
        >
            {header && (
                <Stack.Trigger asChild>
                    <LineWrapper line={substack.lines[0]} />
                </Stack.Trigger>
            )}
            <Stack.Content className="text-sm" header={!!header}>
                {!isHistogram &&
                    contentLines.map((l, i) => {
                        return <LineWrapper key={i} line={l} />;
                    })}
                {isHistogram && (
                    <HistogramWrapper
                        swatches={contentLines.flatMap((l) => l.swatch ?? [])}
                    />
                )}
            </Stack.Content>
        </Stack>
    );
};

const colorInterpolator = interpolateRgbBasis([
    '#fff',
    colors.amber[20],
    colors.violet[80],
]);

const HistogramWrapper = ({ swatches }: { swatches: SwatchLineProto[] }) => {
    const { state } = useStackContext();

    // @ts-expect-error: mismatch between current collection and new typed BE.
    const matchedCondition = state.stack?.proto?.bucketed.find((b: $FixMe) => {
        console.log({ b });
        return b.condition.indices.every((index: number, i: number) => {
            const chip = state.choiceChips[index];
            console.log({ chip });
            if (!chip) return false;
            return chip.value === b.condition.values[i];
        });
    });

    const buckets = matchedCondition?.buckets ?? [];

    const data = useMemo(() => {
        return swatches.flatMap((swatch) => {
            if (swatch.index === -1) return [];
            return {
                index: swatch.index,
                label: swatch.label?.value ?? '',
                count: buckets?.[swatch.index]?.ids
                    ? // this is probably wrong, but just to get a value to show
                      buckets?.[swatch.index]?.ids.reduce(
                          (acc: number, curr: { ids: Array<$FixMe> }) =>
                              acc + curr.ids.length,
                          0
                      ) ?? 0
                    : 0,
            };
        });
    }, [swatches, buckets]);

    const histogramColorScale = useMemo(
        () =>
            scaleOrdinal({
                domain: data.map((d) => `${d.index}`),
                range: data.map((_, i) => colorInterpolator(i / data.length)),
            }),
        [data]
    );
    return (
        <Histogram
            data={data}
            label={(d) => d.label}
            bucket={(d) => d.index.toString()}
            value={(d) => d.count}
            color={(d) => histogramColorScale(`${d.index}`)}
        />
    );
};

export const HeaderWrapper = ({ header }: { header: HeaderLineProto }) => {
    const setAppAtom = useSetAtom(appAtom);
    const {
        state: { stack },
    } = useStackContext();
    const [sharePopoverOpen, setSharePopoverOpen] = useState(false);

    return (
        <Header>
            {header.title && (
                <Header.Label>
                    <AtomWrapper atom={header.title} />
                </Header.Label>
            )}
            <Header.Actions
                close={header.close}
                share={header.share}
                slotProps={{
                    share: {
                        popover: {
                            open: sharePopoverOpen,
                            onOpenChange: setSharePopoverOpen,
                            content: 'Copied to clipboard',
                        },
                        onClick: async (evt) => {
                            evt.preventDefault();
                            evt.stopPropagation();
                            navigator.clipboard
                                .writeText(header?.title?.value ?? '')
                                .then(() => {
                                    setSharePopoverOpen(true);
                                })
                                .catch((err) => {
                                    console.error(
                                        'Failed to copy to clipboard',
                                        err
                                    );
                                });
                        },
                    },
                    close: {
                        onClick: (evt) => {
                            evt.preventDefault();
                            evt.stopPropagation();
                            if (!stack?.id) return;
                            setAppAtom((draft) => {
                                draft.stacks = omit(draft.stacks, stack.id);
                            });
                        },
                    },
                }}
            />
        </Header>
    );
};

export const LineWrapper = ({ line }: { line: LineProto }) => {
    const clickable =
        line.value?.clickExpression ?? line.action?.clickExpression;
    const Wrapper = clickable ? Line.Button : React.Fragment;
    const stack = useStackContext();
    const [app, setApp] = useAtom(appAtom);
    const { [stack.state.mapId]: map } = useMap();

    const { data, refetch } = useQuery({
        queryKey: ['stack', JSON.stringify(clickable)],
        queryFn: () => {
            if (
                !app.startup?.session ||
                !map?.getCenter() ||
                map?.getZoom() === undefined
            ) {
                return null;
            }
            return fetchB6('stack', {
                context: app.startup?.context,
                root: undefined,
                expression: '',
                locked: true,
                logEvent: 'oc',
                logMapCenter: {
                    latE7: Math.round(map.getCenter().lat * 1e7),
                    lngE7: Math.round(map.getCenter().lng * 1e7),
                },
                logMapZoom: map.getZoom(),
                node: clickable,
                session: app.startup?.session,
            }).then((res) => res.json() as Promise<StackResponse>);
        },
        enabled: false,
    });

    useEffect(() => {
        if (data) {
            setApp((draft) => {
                draft.stacks[data.proto.expression] = {
                    proto: data.proto,
                    docked: !!stack.state.stack?.docked,
                    id: data.proto.expression,
                };
            });
        }
    }, [data]);

    return (
        <LineContextProvider line={line}>
            <Line>
                <Wrapper
                    {...(clickable && {
                        onClick: (e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            refetch();
                        },
                    })}
                >
                    {line.header && <HeaderWrapper header={line.header} />}
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
                    {line.choice && <ChoiceWrapper choice={line.choice} />}
                </Wrapper>
            </Line>
        </LineContextProvider>
    );
};

export const ChoiceWrapper = ({
    choice,
}: {
    choice: {
        chips: AtomProto[];
        label: ChoiceProto['label'];
    };
}) => {
    const stack = useStackContext();
    return (
        <>
            {choice.label && <AtomWrapper atom={choice.label} />}
            {choice.chips &&
                choice.chips.map(({ chip }) => {
                    if (!chip) return null;
                    const stackChip = stack.state.choiceChips[chip.index ?? 0];
                    if (stackChip) {
                        return (
                            <ChipWrapper
                                key={chip.index}
                                chip={stackChip}
                                onChange={(value: number) =>
                                    stack.setChoiceChipValue(
                                        chip.index ?? 0, // same issue with the 0 index being undefined, maybe we should add zod to parse this values beforehand or fix in BE.
                                        value
                                    )
                                }
                            />
                        );
                    }
                })}
        </>
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
        return <Line.Value className="text-sm">{atom.value}</Line.Value>;
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
