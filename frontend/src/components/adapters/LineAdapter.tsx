import { appAtom } from '@/atoms/app';
import { Line } from '@/components/system/Line';
import { fetchB6 } from '@/lib/b6';
import { LineContextProvider } from '@/lib/context/line';
import { useStackContext } from '@/lib/context/stack';
import { LineProto, TagsLineProto } from '@/types/generated/ui';
import { StackResponse } from '@/types/stack';
import { useQuery } from '@tanstack/react-query';
import { useAtom } from 'jotai';
import React, { useEffect } from 'react';
import { useMap } from 'react-map-gl/maplibre';
import { TooltipOverflow } from '../system/Tooltip';
import { AtomAdapter } from './AtomAdapter';
import { ChoiceAdapter } from './ChoiceAdapter';
import { HeaderAdapter } from './HeaderAdapter';
import { ShellAdapter } from './ShellAdapter';

export const LineAdapter = ({ line }: { line: LineProto }) => {
    const clickable =
        line.value?.clickExpression ?? line.action?.clickExpression;
    const Wrapper = clickable ? Line.Button : React.Fragment;
    const stack = useStackContext();
    const [app, setApp] = useAtom(appAtom);
    const { [stack.state.mapId]: map } = useMap();

    const { data, refetch } = useQuery({
        queryKey: ['stack-line', JSON.stringify(clickable)],
        queryFn: () => {
            if (
                !app.startup?.session ||
                !map?.getCenter() ||
                map?.getZoom() === undefined
            ) {
                return null;
            }
            return fetchB6('stack', {
                root: undefined,
                expression: '',
                locked: true,
                logEvent: 'oc',
                logMapCenter: {
                    latE7: Math.round(map.getCenter().lat * 1e7),
                    lngE7: Math.round(map.getCenter().lng * 1e7),
                },
                logMapZoom: map.getZoom(),
                node: clickable,
                session: app.startup?.session,
            }).then((res) => res.json() as Promise<StackResponse>);
        },
        enabled: false,
    });

    useEffect(() => {
        if (data) {
            setApp((draft) => {
                draft.stacks[data.proto.expression] = {
                    proto: data.proto,
                    docked: !!stack.state.stack?.docked,
                    id: data.proto.expression,
                };
            });
        }
    }, [data]);

    return (
        <LineContextProvider line={line}>
            <Line>
                <Wrapper
                    {...(clickable && {
                        onClick: (e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            refetch();
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
                                {line.leftRightValue.left.map(({ atom }, i) => {
                                    if (!atom) return null;
                                    return <AtomAdapter key={i} atom={atom} />;
                                })}
                            </div>
                            {line.leftRightValue.right?.atom && (
                                <div className="flex items-center gap-1">
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
                        <span className="expression">
                            {line.expression.expression}
                        </span>
                    )}
                    {line.tags && <Tags tagLine={line.tags} />}
                </Wrapper>
            </Line>
        </LineContextProvider>
    );
};

const Tags = ({ tagLine }: { tagLine: TagsLineProto }) => {
    return (
        <div className="tag w-full text-sm ">
            {tagLine.tags.map((tag) => {
                return (
                    <div className="flex gap-4 justify-between border-b border-b-transparent  hover:border-b-graphite-30 transition-colors ">
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
