import { Cross1Icon, ReaderIcon } from '@radix-ui/react-icons';
import { HTMLMotionProps, motion } from 'framer-motion';
import { isObject, isUndefined } from 'lodash';
import React, {
    HTMLAttributes,
    useCallback,
    useDeferredValue,
    useEffect,
    useState,
} from 'react';
import { twMerge } from 'tailwind-merge';
import { Tab as TabT } from '../stores/tabs';

const Root = React.forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
    ({ children, ...props }, forwardedRef) => {
        return (
            <div {...props} ref={forwardedRef}>
                {children}
            </div>
        );
    }
);

const Menu = React.forwardRef<
    HTMLDivElement,
    HTMLAttributes<HTMLDivElement> & { splitScreen: boolean }
>(({ splitScreen, children, ...props }, forwardedRef) => {
    return (
        <div
            {...props}
            className={twMerge(
                'w-full px-1 pt-2 z-10 -mb-[1px]',
                props.className
            )}
            ref={forwardedRef}
        >
            <div
                className={twMerge(
                    'grid grid-cols-1',
                    splitScreen && 'grid-cols-2'
                )}
            >
                {children}
            </div>
        </div>
    );
});

const Content = React.forwardRef<
    HTMLDivElement,
    HTMLMotionProps<'div'> & HTMLAttributes<HTMLDivElement>
>(({ children, ...props }, forwardedRef) => {
    return (
        <motion.div
            {...props}
            className={twMerge('flex-grow', props.className)}
            ref={forwardedRef}
        >
            {children}
        </motion.div>
    );
});

const Item = React.forwardRef<
    HTMLDivElement,
    HTMLMotionProps<'div'> &
        HTMLAttributes<HTMLDivElement> & {
            side: 'left' | 'right';
            splitScreen: boolean;
        }
>(({ side, splitScreen, children, ...props }, forwardedRef) => {
    return (
        <motion.div
            {...props}
            transition={{
                duration: 0.1618,
                ...props.transition,
            }}
            initial={
                props.initial || isUndefined(props.initial)
                    ? {
                          opacity: 0,
                          x: side === 'right' ? 100 : -100,
                          ...(isObject(props.initial) ? props.initial : {}),
                      }
                    : props.initial
            }
            animate={
                props.animate || isUndefined(props.animate)
                    ? {
                          width: splitScreen ? '50%' : '100%',
                          opacity: 1,
                          x: 0,
                          ...(isObject(props.animate) ? props.animate : {}),
                      }
                    : props.animate
            }
            className={twMerge(
                'h-full border border-x-graphite-40 border-t-graphite-40 border-t bg-graphite-30 relative',
                side === 'right' &&
                    'border-x-rose-40 border-t-rose-40 bg-rose-30',
                splitScreen && 'w-1/2 inline-block',

                props.className
            )}
            ref={forwardedRef}
        >
            <div
                className={twMerge(
                    'h-full w-full relative border-2 border-graphite-30 rounded-lg',
                    side === 'right' && 'border-rose-30'
                )}
            />
            {children}
        </motion.div>
    );
});

const Button = ({
    tab,
    active = false,
    onClose,
    onClick,
    onValueChange,
    ...props
}: {
    tab: TabT;
    side?: 'left' | 'right';
    editable?: boolean;
    closable?: boolean;
    active?: boolean;
    onClose?: (tabId: TabT['id']) => void;
    onClick?: (tabId: TabT['id']) => void;
    onValueChange?: (tabId: TabT['id'], value: string) => void;
} & Omit<HTMLMotionProps<'div'>, 'onClick'>) => {
    const [inputValue, setInputValue] = useState(tab.properties.name);
    const deferredInput = useDeferredValue(inputValue);
    const [isHovered, setIsHovered] = useState(false);

    useEffect(() => {
        if (onValueChange) {
            onValueChange(tab.id, deferredInput);
        }
    }, [deferredInput, onValueChange]);

    const handleInputChange = (evt: React.ChangeEvent<HTMLInputElement>) => {
        setInputValue(evt.target.value);
    };

    const handleClose = useCallback(
        (ev: React.MouseEvent<HTMLButtonElement>) => {
            ev.stopPropagation();
            if (onClose) {
                onClose(tab.id);
            }
        },
        [onClose, tab]
    );

    const handleClick = () => {
        if (onClick) {
            onClick(tab.id);
        }
    };

    return (
        <motion.div
            {...props}
            className={twMerge(
                'text-sm w-fit border-b-2  flex gap-2  items-center transition-colors bg-graphite-20 rounded rounded-b-none border  border-graphite-40 px-2 py-1',
                tab.side === 'right' && 'bg-rose-20 border-rose-40',
                active &&
                    (tab.side === 'right'
                        ? 'border-b-rose-30'
                        : 'border-b-graphite-30'),
                tab.side === 'right'
                    ? 'hover:bg-rose-30'
                    : 'hover:bg-graphite-30',
                active &&
                    (tab.side === 'right' ? 'bg-rose-30' : 'bg-graphite-30'),
                props.className
            )}
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
            onClick={handleClick}
        >
            <ReaderIcon />
            {tab.properties.editable && active ? (
                <input
                    onChange={handleInputChange}
                    disabled={!tab.properties.editable || !active}
                    className="bg-transparent border-none text-sm focus:outline-none focus:text-graphite-80 transition-colors  caret-rose-60 "
                    value={tab.properties.name}
                />
            ) : (
                <span className=" cursor-pointer">{tab.properties.name}</span>
            )}

            {tab.properties.closable && (
                <div className="w-4 flex">
                    <button
                        aria-label="close tab "
                        onClick={handleClose}
                        className={twMerge(
                            'text-rose-70 hover:text-rose-90 transition-colors',
                            !isHovered && ' hidden ',
                            isHovered && ' visible'
                        )}
                    >
                        <Cross1Icon />
                    </button>
                </div>
            )}
        </motion.div>
    );
};

export const Tabs = Object.assign(Root, {
    Menu,
    Button,
    Content,
    Item,
});
