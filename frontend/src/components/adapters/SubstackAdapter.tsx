import { SubstackProto } from '@/types/generated/ui';
import { useState } from 'react';
import { LineWrapper } from '../Renderer';
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
                    <LineAdapter line={substack.lines[0]} />
                </Stack.Trigger>
            )}
            <Stack.Content className="text-sm" header={!!header}>
                {!isHistogram &&
                    contentLines.map((l, i) => {
                        return <LineWrapper key={i} line={l} />;
                    })}
                {isHistogram && (
                    <HistogramAdaptor
                        swatches={contentLines.flatMap((l) => l.swatch ?? [])}
                    />
                )}
            </Stack.Content>
        </Stack>
    );
};
