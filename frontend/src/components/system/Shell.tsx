import { CaretRightIcon } from '@radix-ui/react-icons';
import { Command } from 'cmdk';
import { isNil } from 'lodash';
import { QuickScore } from 'quick-score';
import { useEffect, useMemo, useRef, useState } from 'react';
import { twMerge } from 'tailwind-merge';

import { Line } from '@/components/system/Line';
import { getWordAt, highlighted } from '@/utils/text';

import './Shell.css';

type FunctionB6 = {
    id: string;
    description?: string;
};

/**
 * A Shell component that can be used to execute B6 functions.
 * @TODO: Review the implementation depending on the api to execute functions.
 */
export function Shell({
    onSubmit,
    functions,
    className,
    placeholder,
}: {
    /** The list of functions that can be executed. */
    functions: FunctionB6[];
    /** Optional handler for the submit event. */
    onSubmit?: (expression: string) => void;
    className?: string;
    placeholder?: string;
}) {
    const inputRef = useRef<HTMLInputElement>(null);
    const keywordsRef = useRef<HTMLDivElement>(null);
    const [currentWord, setCurrentWord] = useState<{
        word: string;
        pos: number;
    } | null>(null);
    const [scrollValue, setScrollValue] = useState(0);
    const [input, setInput] = useState('');

    const matcher = useMemo(() => {
        const qs = new QuickScore(functions, ['id', 'description']);
        return qs;
    }, [functions]);

    const functionResults = useMemo(() => {
        if (!currentWord) return [];
        return matcher.search(currentWord.word);
    }, [currentWord, matcher]);

    useEffect(() => {
        if (keywordsRef.current) {
            keywordsRef.current.scrollLeft = scrollValue;
        }
    }, [scrollValue]);

    return (
        <Command
            label="Shell"
            /* Current filtering logic is a bit naive, but works for now. In the future we can
            integrate match-sorter https://github.com/kentcdodds/match-sorter */
            shouldFilter={false}
            className={twMerge('shell w-full', className)}
        >
            <Line className="flex gap-2 bg-ultramarine-10 hover:bg-ultramarine-10 w-full ">
                <span className="text-ultramarine-70 "> b6</span>
                <div className="relative flex-grow">
                    <div ref={keywordsRef} className="keywords">
                        {input.split(' ').map((word, index) => {
                            const isFunction = functions.some(
                                (f) => f.id === word
                            );
                            const isPipe = word === '|';
                            return (
                                <span
                                    key={index}
                                    className={twMerge(
                                        'text-graphite-70',
                                        isFunction &&
                                            'text-ultramarine-70 italic',
                                        isPipe && ' text-graphite-100 '
                                    )}
                                >
                                    {word}&nbsp;
                                </span>
                            );
                        })}
                    </div>
                    <Command.Input
                        placeholder={placeholder}
                        ref={inputRef}
                        value={input}
                        onSubmit={() => {
                            if (onSubmit) {
                                onSubmit(input);
                                setInput('');
                            }
                        }}
                        className={twMerge(
                            'input caret-ultramarine-60 bg-transparent text-transparent focus:outline-none'
                        )}
                        onKeyDown={(evt) => {
                            if (
                                evt.key === 'Enter' &&
                                functionResults.length === 0 &&
                                input !== '' &&
                                onSubmit
                            ) {
                                onSubmit(input);
                                setInput('');
                            }
                        }}
                        onValueChange={(v) => setInput(v)}
                        onSelect={() => {
                            if (!inputRef.current) return;
                            const s = inputRef.current.selectionEnd;
                            if (isNil(s)) return;
                            const word = getWordAt(input, s);
                            setCurrentWord({
                                word,
                                pos: s,
                            });
                        }}
                        onScroll={(e) => {
                            setScrollValue(e.currentTarget.scrollLeft);
                        }}
                    />
                </div>
            </Line>
            {input !== '' && (
                <Command.List className="[&_.line]:border-t-0  first:border-t first:border-t-graphite-30  max-h-64 overflow-y-auto border-b border-b-graphite-30 ">
                    {functionResults.map((f) => (
                        <Command.Item
                            className="[&_.line]:data-[selected=true]:bg-ultramarine-10 transition-colors [&_.line]:data-[selected=true]:border-l [&_.line]:data-[selected=true]:border-l-ultramarine-60 [&_.line]:last:border-b-0  "
                            key={f.item.id}
                            value={f.item.id}
                            onSelect={() => {
                                const newInput = `${input.slice(
                                    0,
                                    currentWord!.pos - currentWord!.word.length
                                )} ${f.item.id} ${input.slice(
                                    currentWord!.pos
                                )}`;

                                setInput(newInput);
                            }}
                        >
                            <Line className="flex flex-row gap-1 items-center">
                                <CaretRightIcon className=" text-ultramarine-60  shrink-0" />
                                <div className="flex gap-2 items-baseline [&strong]:font-medium">
                                    <span>
                                        {highlighted(f.item.id, f.matches.id)}
                                    </span>

                                    <span className="text-xs text-graphite-60 ">
                                        {f.item?.description &&
                                            highlighted(
                                                f.item.description,
                                                f.matches.description
                                            )}
                                    </span>
                                </div>
                            </Line>
                        </Command.Item>
                    ))}
                </Command.List>
            )}
        </Command>
    );
}
