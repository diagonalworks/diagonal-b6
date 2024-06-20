import { useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';

import { useStackContext } from '@/lib/context/stack';
import { SubstackProto } from '@/types/generated/ui';

import { Stack } from '../system/Stack';
import { HistogramAdaptor } from './HistogramAdapter';
import { LineAdapter } from './LineAdapter';

export const SubstackAdapter = ({
    substack,
    collapsible = false,
    close,
    show,
    analysisTitle,
}: {
    substack: SubstackProto;
    collapsible?: boolean;
    close?: boolean;
    show?: boolean;
    analysisTitle?: string; // @TODO: remove this, it's a hack to get the analysis title
}) => {
    const [open, setOpen] = useState(collapsible ? false : true);
    const { origin } = useStackContext();
    const header = useMemo(() => {
        return substack.lines?.[0].header;
    }, [substack.lines]);

    const contentLines = useMemo(() => {
        return substack.lines?.slice(header || collapsible ? 1 : 0) ?? [];
    }, [substack.lines, header, collapsible]);

    const isHistogram =
        contentLines.length > 1 &&
        (contentLines.every((line) => line.swatch) ||
            contentLines.every((line) => line.histogramBar));

    const histogramProps = useMemo(() => {
        if (isHistogram) {
            // @TODO: find more robust way to access origin histogram
            const originHistogram = origin?.data?.proto.stack?.substacks?.[0];
            return {
                type: (contentLines[0].swatch ? 'swatch' : 'histogram') as
                    | 'swatch'
                    | 'histogram',
                bars: contentLines.flatMap((l) => l.histogramBar ?? []),
                swatches: contentLines.flatMap((l) => l.swatch ?? []),
                chartLabel: analysisTitle,
                origin: originHistogram
                    ? {
                          bars:
                              originHistogram.lines
                                  ?.slice(header ? 1 : 0)
                                  ?.flatMap((l) => l.histogramBar ?? []) ?? [],
                          swatches:
                              originHistogram.lines
                                  ?.slice(header ? 1 : 0)
                                  ?.flatMap((l) => l.swatch ?? []) ?? [],
                      }
                    : undefined,
            };
        }
        return null;
    }, [contentLines]);

    if (!substack.lines) return null;

    return (
        <Stack
            collapsible={collapsible}
            open={open}
            onOpenChange={setOpen}
            className="w-full"
        >
            {(header || collapsible) && (
                <Stack.Trigger
                    className={twMerge(
                        'w-full border-b border-graphite-30 cursor-pointer',
                        open ? 'border-b-0' : ''
                    )}
                >
                    <LineAdapter
                        line={substack.lines[0]}
                        //changeable={changeable}
                    />
                </Stack.Trigger>
            )}

            <Stack.Content
                className="text-sm max-h-80 w-full"
                header={!!header}
            >
                {!isHistogram &&
                    contentLines.map((l, i) => {
                        return (
                            <LineAdapter
                                key={i}
                                line={l}
                                actions={{
                                    close: close && i === 0,
                                    show: show && i === 0,
                                }}
                            />
                        );
                    })}
                {isHistogram && histogramProps && (
                    <HistogramAdaptor {...histogramProps} />
                )}
            </Stack.Content>
        </Stack>
    );
};
