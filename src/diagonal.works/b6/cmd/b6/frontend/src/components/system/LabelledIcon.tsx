import React from 'react';
import { twMerge } from 'tailwind-merge';

export interface LabelledIconProps
    extends React.HTMLAttributes<HTMLDivElement> {
    slots: {
        icon: React.ReactElement;
    };
}

export const LabelledIcon = (props: LabelledIconProps) => {
    return (
        <div
            className={twMerge(
                'flex gap-2 items-center text-graphite-100 w-fit ',
                props.className
            )}
        >
            {React.cloneElement(props.slots.icon, {
                className: twMerge(
                    'fill-graphite-80',
                    props.slots.icon.props.className
                ),
                ...props.slots.icon.props,
            })}
            {props.children}
        </div>
    );
};
