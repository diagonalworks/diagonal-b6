import { SubstackProto } from '@/types/generated/ui';
import { useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Stack } from '../system/Stack';
import { HistogramAdaptor } from './HistogramAdapter';
import { LineAdapter } from './LineAdapter';

export const SubstackAdapter = ({
    substack,
    collapsible = false,
    close,
    origin,
}: {
    substack: SubstackProto;
    collapsible?: boolean;
    close?: boolean;
    origin?: SubstackProto;
}) => {
    const [open, setOpen] = useState(collapsible ? false : true);

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
            return {
                type: (contentLines[0].swatch ? 'swatch' : 'histogram') as
                    | 'swatch'
                    | 'histogram',
                bars: contentLines.flatMap((l) => l.histogramBar ?? []),
                swatches: contentLines.flatMap((l) => l.swatch ?? []),
                origin: origin
                    ? {
                          bars:
                              origin.lines
                                  ?.slice(header ? 1 : 0)
                                  ?.flatMap((l) => l.histogramBar ?? []) ?? [],
                          swatches:
                              origin.lines
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
        <Stack collapsible={collapsible} open={open} onOpenChange={setOpen}>
            {(header || collapsible) && (
                <Stack.Trigger
                    className={twMerge(
                        'border-b border-graphite-30 cursor-pointer',
                        open ? 'border-b-0' : ''
                    )}
                >
                    <LineAdapter
                        line={substack.lines[0]}
                        //changeable={changeable}
                    />
                </Stack.Trigger>
            )}

            <Stack.Content className="text-sm max-h-80 " header={!!header}>
                {!isHistogram &&
                    contentLines.map((l, i) => {
                        return (
                            <LineAdapter
                                key={i}
                                line={l}
                                close={close && i === 0}
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
