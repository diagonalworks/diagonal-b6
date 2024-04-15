// Storage object for jotai atomWithStorage that stores the value of the atom in the URL search params
export const urlSearchParamsStorage = <T>({
    serialize = (value) => String(value),
    deserialize = (value) => value as unknown as T,
}: {
    serialize?: (value: T) => string;
    deserialize?: (value: string) => T;
}) => {
    return {
        getItem: (key: string, initialValue: T) => {
            const params = new URLSearchParams(window.location.search);
            const value = params.get(key);
            return value ? deserialize(value) : initialValue;
        },
        setItem: (key: string, newValue: T) => {
            const serialized = serialize(newValue);
            const params = new URLSearchParams(window.location.search);
            params.set(key, serialized);
            window.history.replaceState(
                {},
                '',
                `${window.location.pathname}?${params}`
            );
        },
        removeItem: (key: string) => {
            const params = new URLSearchParams(window.location.search);
            params.delete(key);
            window.history.replaceState(
                {},
                '',
                `${window.location.pathname}?${params}`
            );
        },
    };
};
