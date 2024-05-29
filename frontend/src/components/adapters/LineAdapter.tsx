import { Line } from '@/components/system/Line';
import { LineContextProvider } from '@/lib/context/line';
import { useStackContext } from '@/lib/context/stack';
import { LineProto, TagsLineProto } from '@/types/generated/ui';
import { Cross1Icon } from '@radix-ui/react-icons';
import React from 'react';
import { twMerge } from 'tailwind-merge';
import { IconButton } from '../system/IconButton';
import { TooltipOverflow } from '../system/Tooltip';
import { AtomAdapter } from './AtomAdapter';
import { ChoiceAdapter } from './ChoiceAdapter';
import { HeaderAdapter } from './HeaderAdapter';
import { ShellAdapter } from './ShellAdapter';

export const LineAdapter = ({
    line,
    close,
}: {
    line: LineProto;
    close?: boolean;
}) => {
    const clickable =
        line.value?.clickExpression ?? line.action?.clickExpression;
    const Wrapper = clickable ? Line.Button : React.Fragment;
    const { close: closeFn, evaluateNode } = useStackContext();

    const handleLineClick = () => {
        if (!clickable) return;
        evaluateNode(clickable);
    };

    return (
        <LineContextProvider line={line}>
            <Line
                className={twMerge(
                    'flex justify-between',
                    line.error && 'bg-red-20 text-red-70 hover:bg-red-20',
                    line.expression && 'bg-graphite-10 '
                )}
            >
                <Wrapper
                    {...(clickable && {
                        onClick: (e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            handleLineClick();
                        },
                    })}
                >
                    {line.header && <HeaderAdapter header={line.header} />}
                    {/* {line.choice && <SelectWrapper choice={line.choice} />} */}
                    {line.value && line.value.atom && (
                        <AtomAdapter atom={line.value.atom} />
                    )}
                    {line.leftRightValue && (
                        <div className="justify-between flex items-center w-full">
                            <div className="flex items-center gap-2 w-11/12 flex-grow-0">
                                {line.leftRightValue.left?.map(
                                    ({ atom }, i) => {
                                        if (!atom) return null;
                                        return (
                                            <AtomAdapter key={i} atom={atom} />
                                        );
                                    }
                                )}
                            </div>
                            {line.leftRightValue.right?.atom && (
                                <div className="flex items-center gap-1 text-ultramarine-50">
                                    <AtomAdapter
                                        atom={line.leftRightValue.right.atom}
                                    />
                                </div>
                            )}
                        </div>
                    )}
                    {line.choice && <ChoiceAdapter choice={line.choice} />}
                    {line.shell && <ShellAdapter shell={line.shell} />}
                    {line.expression && (
                        <span className="expression ">
                            {line.expression.expression}
                        </span>
                    )}
                    {line.error && (
                        <span className="error">
                            <span className=" font-medium mr-1">Error:</span>
                            {line.error.error}
                        </span>
                    )}
                    {line.tags && <Tags tagLine={line.tags} />}
                    {close && (
                        <IconButton
                            onClick={(e) => {
                                e.preventDefault();
                                e.stopPropagation();
                                closeFn();
                            }}
                            className="close"
                        >
                            <Cross1Icon />
                        </IconButton>
                    )}
                </Wrapper>
            </Line>
        </LineContextProvider>
    );
};

const Tags = ({ tagLine }: { tagLine: TagsLineProto }) => {
    return (
        <div className="tag w-full text-sm ">
            {tagLine.tags?.map((tag, i) => {
                return (
                    <div
                        key={i}
                        className="flex gap-4 justify-between border-b border-b-transparent  hover:border-b-graphite-30 transition-colors "
                    >
                        <div className="flex gap-2 text-graphite-80 ">
                            <span className=" min-w-2 italic">
                                {tag.prefix}
                            </span>
                            <span className="font-medium">{tag.key}</span>
                        </div>
                        <div className=" max-w-1/2 text-right font-medium">
                            <TooltipOverflow>{tag.value}</TooltipOverflow>
                        </div>
                    </div>
                );
            })}
        </div>
    );
};
