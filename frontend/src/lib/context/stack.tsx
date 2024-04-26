import { AppStore } from '@/atoms/app';
import { Chip } from '@/types/stack';
import { isUndefined } from 'lodash';
import {
    PropsWithChildren,
    createContext,
    useCallback,
    useContext,
    useMemo,
} from 'react';
import { Updater, useImmer } from 'use-immer';

const StackContext = createContext<{
    state: StackStore;
    setState: Updater<StackStore>;
    setChoiceChipValue: (index: number, value: number) => void;
}>({
    state: {
        mapId: 'baseline',
        choiceChips: {},
    },
    setState: () => {},
    setChoiceChipValue: () => {},
});

export type StackStore = {
    mapId: string;
    stack?: AppStore['stacks'][string];
    choiceChips: Record<number, Chip>;
};

export const StackContextProvider = ({
    stack,
    mapId,
    children,
}: {
    stack: AppStore['stacks'][string];
    mapId: string;
} & PropsWithChildren) => {
    const choiceChips = useMemo(() => {
        const chips: Record<number, Chip> = {};
        // Which substack is the choice line in? should substacks have their own context?
        const allLines =
            stack.proto.stack?.substacks.flatMap(
                (substack) => substack.lines
            ) ?? [];
        const choiceLines = allLines.flatMap((line) => line?.choice ?? []);

        choiceLines.forEach((line) => {
            line.chips.forEach((atom) => {
                if (isUndefined(atom.chip?.index)) {
                    console.warn(`Chip index is undefined`, { line, atom });
                }
                const chipIndex = atom.chip?.index ?? 0; // unsafe fallback
                chips[chipIndex] = {
                    atom: {
                        labels: atom.chip?.labels ?? [],
                        index: chipIndex,
                    },
                    value: 0,
                };
            });
        });
        return chips;
    }, [stack]);

    const [state, setState] = useImmer<StackStore>({
        mapId,
        stack,
        choiceChips,
    });

    const setChoiceChipValue = useCallback(
        (index: number, value: number) => {
            setState((draft) => {
                if (!draft.choiceChips[index]) return;
                draft.choiceChips[index].value = value;
            });
        },
        [setState]
    );

    const stackContextStoreData = useMemo(() => {
        return {
            state,
            setState,
            setChoiceChipValue,
        };
    }, [state, setState, setChoiceChipValue]);

    return (
        <StackContext.Provider value={stackContextStoreData}>
            {children}
        </StackContext.Provider>
    );
};

// eslint-disable-next-line react-refresh/only-export-components
export const useStackContext = () => useContext(StackContext);
