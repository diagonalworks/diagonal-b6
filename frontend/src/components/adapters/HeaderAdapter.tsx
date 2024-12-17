import { useState } from 'react';

import { AtomAdapter } from '@/components/adapters/AtomAdapter';
import { Header } from '@/components/system/Header';
import { useStackContext } from '@/lib/context/stack';
import { HeaderLineProto } from '@/types/generated/ui';

export const HeaderAdapter = ({ header }: { header: HeaderLineProto }) => {
    const [sharePopoverOpen, setSharePopoverOpen] = useState(false);
    const { close, evaluateNode, data } = useStackContext();

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
                target={header.target}
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
                    target: {
                        onClick: (evt) => {
                            evt.preventDefault();
                            evt.stopPropagation();
                            if( data?.proto?.node ) {
                                // Evaluate the node; but don't show it on the
                                // list of outliners; and also force a (query)
                                // cache refresh so that it centers on that
                                // point.
                                evaluateNode(data.proto.node, false, true);
                            }
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
