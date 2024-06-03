import { findAtoms } from '@/lib/atoms';
import { OutlinerSpec } from '@/stores/outliners';
import { StackProto } from '@/types/generated/ui';
import { MinusIcon, PlusIcon } from '@radix-ui/react-icons';
import { isEqual } from 'lodash';
import { PropsWithChildren } from 'react';
import { twMerge } from 'tailwind-merge';
import { Change, useChangesStore } from '../stores/changes';

export const OutlinerChangeWrapper = ({
    id,
    outliner,
    stack,
    children,
}: {
    id: Change['id'];
    outliner: OutlinerSpec;
    stack?: StackProto;
} & PropsWithChildren) => {
    const actions = useChangesStore((state) => state.actions);
    const change = useChangesStore((state) => state.changes[id]);

    const featureId = stack?.id;

    const isInChange = change.spec.features.find((f) =>
        isEqual(f.id, featureId)
    );

    const showChangeElements = featureId && !change.created;

    const labelledIcon = stack?.substacks?.[0]?.lines?.flatMap((l) =>
        findAtoms(l, 'labelledIcon')
    )?.[0]?.labelledIcon;

    if (!showChangeElements) return null;

    return (
        <>
            <div className="flex justify-start">
                <button
                    onClick={() => {
                        const expression = outliner.request?.expression ?? '';

                        if (isInChange) {
                            actions.removeFeature(id, {
                                expression,
                                id: featureId,
                                label: labelledIcon,
                            });
                        }
                        actions.addFeature(id, {
                            expression,
                            id: featureId,
                            label: labelledIcon,
                        });
                    }}
                    className="-mb-[2px] p-2 flex gap-1  items-center text-xs  text-rose-90 rounded-t border-b-0 bg-rose-40 hover:bg-rose-30 border border-rose-50"
                >
                    {isInChange ? (
                        <>
                            <MinusIcon /> remove from change
                        </>
                    ) : (
                        <>
                            <PlusIcon /> add to change
                        </>
                    )}
                </button>
            </div>
            <div
                className={twMerge(
                    'stack-wrapper',
                    showChangeElements && ' border border-rose-50 rounded'
                )}
            >
                {children}
            </div>
        </>
    );
};
