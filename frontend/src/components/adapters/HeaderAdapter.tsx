import { Header } from '@/components/system/Header';
import { useOutlinerContext } from '@/lib/context/outliner';
import { HeaderLineProto } from '@/types/generated/ui';
import { useState } from 'react';
import { AtomAdapter } from './AtomAdapter';

export const HeaderAdapter = ({ header }: { header: HeaderLineProto }) => {
    const [sharePopoverOpen, setSharePopoverOpen] = useState(false);
    const { close } = useOutlinerContext();

    return (
        <Header>
            {header.title && (
                <Header.Label>
                    <AtomAdapter atom={header.title} />
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
                            close();
                        },
                    },
                }}
            />
        </Header>
    );
};
