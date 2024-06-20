import { isUndefined } from 'lodash';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useMemo,
} from 'react';
import { Updater, useImmer } from 'use-immer';

import { LineProto } from '@/types/generated/ui';
import { Chip } from '@/types/stack';
import { findAtoms } from '@/utils/atoms';

const LineContext = createContext<{
    state: LineStore;
    setState: Updater<LineStore>;
    setChipValue: (index: number, value: number) => void;
}>({
    state: { line: {}, chips: {} },
    setState: () => {},
    setChipValue: () => {},
});

export type LineStore = {
    line: LineProto;
    chips: Record<number, Chip>;
};

export const LineContextProvider = ({
    line,
    children,
}: {
    line: LineProto;
} & PropsWithChildren) => {
    const chips = useMemo(() => {
        const chipMap: LineStore['chips'] = {};

        for (const atom of findAtoms(line, 'chip')) {
            if (atom.chip) {
                if (!atom.chip) {
                    console.warn(`Chip is undefined`, { line, atom });
                    continue;
                }

                if (isUndefined(atom.chip.index)) {
                    console.warn(`Chip index is undefined`, { line, atom });
                    continue;
                }

                chipMap[atom.chip.index] = {
                    atom: {
                        labels: atom.chip.labels ?? [],
                        /* // unsafe fallback but looks like 0 is being considered as undefined and not coming through */
                        index: atom.chip.index ?? 0,
                    },
                    value: 0,
                };
            }
        }

        if (line.choice && line.choice.chips) {
            line.choice.chips.forEach((atom, i) => {
                if (isUndefined(atom.chip?.index)) {
                    console.warn(`Chip index is undefined`, { line, atom });
                }
                chipMap[i] = {
                    atom: {
                        labels: atom.chip?.labels ?? [],
                        index: atom.chip?.index ?? 0, // unsafe fallback
                    },
                    value: 0,
                };
            });
        }

        return chipMap;
    }, [line]);

    const [state, setState] = useImmer<LineStore>({
        line,
        chips,
    });

    const setChipValue = useCallback(
        (index: number, value: number) => {
            setState((draft) => {
                if (!draft.chips[index]) return;
                draft.chips[index].value = value;
            });
        },
        [setState]
    );

    const lineContextStoreData = useMemo(() => {
        return {
            state,
            setState,
            setChipValue,
        };
    }, [state, setState, setChipValue]);
    return (
        <LineContext.Provider value={lineContextStoreData}>
            {children}
        </LineContext.Provider>
    );
};

/**
 * Hook for using the line context.
 * This context is used to provide the line and chip data to the children components.
 * @TODO: This needs to be refactored. We're no longer using a context provider for the Outliner,
 * therefore chip data is not provided. Also, we should now be using context providers mostly for
 * dependency injection and zustand for state management.
 */
export const useLineContext = () => useContext(LineContext);
