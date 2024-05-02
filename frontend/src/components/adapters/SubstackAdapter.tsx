import { SubstackProto } from '@/types/generated/ui';
import { useMemo, useState } from 'react';
import { Stack } from '../system/Stack';
import { HistogramAdaptor } from './HistogramAdapter';
import { LineAdapter } from './LineAdapter';

export const SubstackAdapter = ({
    substack,
    collapsible = false,
}: {
    substack: SubstackProto;
    collapsible?: boolean;
}) => {
    const [open, setOpen] = useState(collapsible ? false : true);

    const header = useMemo(() => {
        return substack.lines?.[0].header;
    }, [substack.lines]);

    const contentLines = useMemo(() => {
        return substack.lines?.slice(header ? 1 : 0) ?? [];
    }, [substack.lines, header]);

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
            };
        }
        return null;
    }, [contentLines]);

    if (!substack.lines) return null;

    return (
        <Stack
            collapsible={substack.collapsable}
            open={open}
            onOpenChange={setOpen}
        >
            {header && (
                <Stack.Trigger asChild>
                    <LineAdapter line={substack.lines[0]} />
                </Stack.Trigger>
            )}
            <Stack.Content className="text-sm max-h-80 " header={!!header}>
                {!isHistogram &&
                    contentLines.map((l, i) => {
                        return <LineAdapter key={i} line={l} />;
                    })}
                {isHistogram && histogramProps && (
                    <HistogramAdaptor {...histogramProps} />
                )}
            </Stack.Content>
        </Stack>
    );
};
