import { CaretRightIcon } from '@radix-ui/react-icons';
import { Command } from 'cmdk';
import { useMemo, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Line } from './Line';

type FunctionB6 = {
    id: string;
    description: string;
};

/**
 * A Shell component that can be used to execute B6 functions.
 * @TODO: Review the implementation depending on the api to execute functions.
 */
export function Shell({
    onSubmit,
    functions,
}: {
    /** The list of functions that can be executed. */
    functions: FunctionB6[];
    /** Optional handler for the submit event. */
    onSubmit?: ({ func, args }: { func: string; args: string }) => void;
}) {
    const [selected, setSelected] = useState<string | null>(null);
    const [input, setInput] = useState('');

    const selectedFunction = useMemo(() => {
        return functions.find((f) => f.id === selected);
    }, [functions, selected]);

    const handleChange = (value: string) => {
        const match = functions.find((f) => f.id === value);
        if (match) {
            setSelected(match.id);
            setInput('');
        } else {
            setInput(value);
        }
    };

    return (
        <Command
            label="Shell"
            /* Current filtering logic is a bit naive, but works for now. In the future we can
            integrate match-sorter https://github.com/kentcdodds/match-sorter */
            filter={(value, search) => (value.includes(search) ? 1 : 0)}
            className="shell w-fit"
        >
            <Line className="flex gap-2 bg-ultramarine-10 hover:bg-ultramarine-10 ">
                <span className="text-ultramarine-70 "> b6</span>
                {selectedFunction && (
                    <span className="text-graphite-100 italic ">
                        {selectedFunction.id}
                    </span>
                )}
                <Command.Input
                    value={input}
                    onSubmit={(evt) => {
                        if (onSubmit) {
                            onSubmit({
                                func: selectedFunction?.id ?? '',
                                args: evt.currentTarget.value,
                            });

                            setSelected(null);
                            setInput('');
                        }
                    }}
                    className={twMerge(
                        'flex-grow caret-ultramarine-60 bg-transparent text-graphite-70 focus:outline-none'
                    )}
                    onKeyDown={(evt) => {
                        if (
                            evt.key === 'Backspace' &&
                            evt.currentTarget.value === ''
                        ) {
                            setSelected(null);
                        }
                        if (evt.key === 'Enter' && selectedFunction) {
                            if (onSubmit) {
                                onSubmit({
                                    func: selectedFunction?.id ?? '',
                                    args: evt.currentTarget.value,
                                });
                            }
                        }
                    }}
                    onValueChange={handleChange}
                />
            </Line>
            {!selected && input !== '' && (
                <Command.List className="[&_.line]:border-t-0 w-80 first:border-t first:border-t-graphite-30  max-h-64 overflow-y-auto border-b border-b-graphite-30 ">
                    <Command.Empty className="text-graphite-60 text-xs">
                        <Line className="border-b-0">No function found</Line>
                    </Command.Empty>
                    {functions.map((f) => (
                        <Command.Item
                            className="[&_.line]:data-[selected=true]:bg-ultramarine-10 transition-colors [&_.line]:data-[selected=true]:border-l [&_.line]:data-[selected=true]:border-l-ultramarine-60 [&_.line]:last:border-b-0  "
                            key={f.id}
                            onSelect={() => {
                                setSelected(f.id);
                                setInput('');
                            }}
                        >
                            <Line className="flex flex-row gap-1 items-start">
                                <CaretRightIcon className=" text-ultramarine-60 mt-1.5 shrink-0" />
                                <div className="flex gap-2 items-baseline">
                                    <span>{f.id}</span>
                                    <span className="text-xs text-graphite-60 ">
                                        {f.description}
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
