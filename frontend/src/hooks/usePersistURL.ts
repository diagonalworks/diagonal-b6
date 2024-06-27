import { isNil } from 'lodash';
import { useEffect } from 'react';
import { StoreApi, UseBoundStore, useStore } from 'zustand';

import { $IntentionalAny } from '@/utils/defs';

export const encodeStateToUrl = <T>(
    state: Partial<T>,
    encode: (state: Partial<T>) => Record<string, string>
) => {
    const params = new URLSearchParams(window.location.search);
    const encoded = encode(state);
    for (const [key, value] of Object.entries(encoded)) {
        params.set(key, value);
    }
    return params.toString();
};

export const decodeUrlToState = <T>(
    search: string,
    decode: (params: Record<string, string>) => (state: T) => T
) => {
    const params = new URLSearchParams(search);
    const decoded = decode(Object.fromEntries(params.entries()));
    return decoded;
};

const debounce = <T extends (...args: $IntentionalAny[]) => $IntentionalAny>(
    func: T,
    wait: number
) => {
    let timeout: NodeJS.Timeout;
    return (...args: Parameters<T>): void => {
        clearTimeout(timeout);
        timeout = setTimeout(() => func(...args), wait);
    };
};

/**
 * Hook for persisting the state of a Zustand store in the URL.
 *
 * @param store - The Zustand store to persist.
 * @param encode - A function to encode the state to a URL query string.
 * @param decode - A function to decode the URL query string to a state.
 * @param debounceWait - The debounce time for updating the URL.
 */
export const usePersistURL = <T>(
    store: StoreApi<T> | UseBoundStore<StoreApi<T>>,
    encode: (state: Partial<T>) => Record<string, string>,
    decode: (
        params: Record<string, string>,
        initial?: boolean
    ) => (state: T) => T,
    debounceWait: number = 100
) => {
    // We need to use the state as a hook dependency, even though we don't use it. Hence, we disable the eslint rule.
    // @ts-expect-error - We need to use the state as a hook dependency, even though we don't use it.
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const state = useStore(store);

    // Initialize state from URL
    useEffect(() => {
        const search = window.location.search;
        const initialState = decodeUrlToState<T>(search, (params) =>
            decode(params, true)
        );
        store.setState(initialState);
    }, [decode]);

    const debouncedUpdateURL = debounce((newState: T) => {
        const queryString = encode(newState);
        const url = new URL(window.location.href);
        const searchParams = new URLSearchParams(url.search);

        Object.entries(queryString).forEach(([key, value]) => {
            if (isNil(value) || value === '') {
                searchParams.delete(key);
            } else {
                searchParams.set(key, value);
            }
        });
        const newParams = new URLSearchParams(searchParams).toString();
        /**
         * we don't encode the URL params because:
         * 1. We want to keep the URL human-readable and editable
         * 2. We want to avoid double encoding issues
         */
        url.search = decodeURIComponent(newParams);
        window.history.pushState({}, '', url.toString());
    }, debounceWait);

    // Subscribe to state changes and update URL
    useEffect(() => {
        const unsubscribe = store.subscribe(debouncedUpdateURL);
        return unsubscribe;
    }, [store, encode]);
};
