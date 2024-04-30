import { appAtom } from '@/atoms/app';
import { Header } from '@/components/system/Header';
import { HeaderLineProto } from '@/types/generated/ui';
import { useSetAtom } from 'jotai';
import { useState } from 'react';
import { AtomAdapter } from './AtomAdapter';

export const HeaderAdapter = ({ header }: { header: HeaderLineProto }) => {
    const setAppAtom = useSetAtom(appAtom);
    //const { stack } = useOutlinerContext();
    const [sharePopoverOpen, setSharePopoverOpen] = useState(false);

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
                            /* if (!stack?.id) return;
                            setAppAtom((draft) => {
                                draft.stacks = omit(draft.stacks, stack.id);
                            }); */
                        },
                    },
                }}
            />
        </Header>
    );
};
