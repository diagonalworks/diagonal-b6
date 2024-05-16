import { urlSearchParamsStorage } from '@/lib/storage';
import { atom } from 'jotai';
import { atomWithStorage } from 'jotai/utils';
import { ViewState } from 'react-map-gl';

const INITIAL_COORDINATES = { latE7: 515361156, lngE7: -1255161 };
const INITIAL_ZOOM = 16;
export const INITIAL_CENTER = {
    lat: INITIAL_COORDINATES.latE7 / 1e7,
    lng: INITIAL_COORDINATES.lngE7 / 1e7,
};

const initialViewParamsFromUrl = () => {
    const params = new URLSearchParams(window.location.search);
    const mapCenter = params.get('ll')?.split(',');
    const zoom = params.get('z');

    const initialViewState = {
        latitude: mapCenter && parseFloat(mapCenter[0]),
        longitude: mapCenter && parseFloat(mapCenter[1]),
        zoom: zoom ? parseInt(zoom) : INITIAL_ZOOM,
    };

    return initialViewState;
};

const { latitude, longitude, zoom } = initialViewParamsFromUrl();

export const zoomStorageAtom = atomWithStorage<number>(
    'z',
    INITIAL_ZOOM,
    urlSearchParamsStorage({}),
    {
        getOnInit: true,
    }
);

export const centerStorageAtom = atomWithStorage<{
    lat?: number;
    lng?: number;
}>(
    'll',
    INITIAL_CENTER,
    urlSearchParamsStorage({
        serialize: (value) =>
            value.lat && value.lng ? `${value.lat},${value.lng}` : '',
        deserialize: (value) => {
            if (!value)
                return {
                    lat: latitude && latitude / 1e7,
                    lng: longitude && longitude / 1e7,
                };
            const [lat, lng] = value.split(',').map(Number);
            return { lat, lng };
        },
    }),
    {
        getOnInit: true,
    }
);

const debouncedCenter = atomWithDebounce(
    {
        lat: latitude,
        lng: longitude,
    },
    200
);
const debouncedZoom = atomWithDebounce(zoom, 200);

const centerAtom = atom({
    lat: latitude,
    lng: longitude,
});
const zoomAtom = atom(zoom);

const centerStorageSetAtom = atom(null, (get, set) => {
    const center = get(debouncedCenter.debouncedValueAtom);
    set(centerStorageAtom, center);
});

const zoomStorageSetAtom = atom(null, (get, set) => {
    const zoom = get(debouncedZoom.debouncedValueAtom);
    set(zoomStorageAtom, zoom);
});

export const viewAtom = atom(
    (get) => {
        return {
            latitude: get(centerAtom).lat,
            longitude: get(centerAtom).lng,
            zoom: get(zoomAtom),
            bearing: 0,
            pitch: 0,
            padding: { top: 0, bottom: 0, left: 0, right: 0 },
        };
    },
    (_, set, view: ViewState) => {
        const center = { lat: view.latitude, lng: view.longitude };
        const zoom = view.zoom;
        set(centerAtom, center);
        set(zoomAtom, zoom);
        set(debouncedCenter.debouncedValueAtom, center);
        set(debouncedZoom.debouncedValueAtom, zoom);
        set(centerStorageSetAtom);
        set(zoomStorageSetAtom);
    }
);

import { SetStateAction } from 'jotai';

export default function atomWithDebounce<T>(
    initialValue: T,
    delayMilliseconds = 500,
    shouldDebounceOnReset = false
) {
    const prevTimeoutAtom = atom<ReturnType<typeof setTimeout> | undefined>(
        undefined
    );

    // DO NOT EXPORT currentValueAtom as using this atom to set state can cause
    // inconsistent state between currentValueAtom and debouncedValueAtom
    const _currentValueAtom = atom(initialValue);
    const isDebouncingAtom = atom(false);

    const debouncedValueAtom = atom(
        initialValue,
        (get, set, update: SetStateAction<T>) => {
            clearTimeout(get(prevTimeoutAtom));

            const prevValue = get(_currentValueAtom);
            const nextValue =
                typeof update === 'function'
                    ? (update as (prev: T) => T)(prevValue)
                    : update;

            const onDebounceStart = () => {
                set(_currentValueAtom, nextValue);
                set(isDebouncingAtom, true);
            };

            const onDebounceEnd = () => {
                set(debouncedValueAtom, nextValue);
                set(isDebouncingAtom, false);
            };

            onDebounceStart();

            if (!shouldDebounceOnReset && nextValue === initialValue) {
                onDebounceEnd();
                return;
            }

            const nextTimeoutId = setTimeout(() => {
                onDebounceEnd();
            }, delayMilliseconds);

            // set previous timeout atom in case it needs to get cleared
            set(prevTimeoutAtom, nextTimeoutId);
        }
    );

    // exported atom setter to clear timeout if needed
    const clearTimeoutAtom = atom(null, (get, set) => {
        clearTimeout(get(prevTimeoutAtom));
        set(isDebouncingAtom, false);
    });

    return {
        currentValueAtom: atom((get) => get(_currentValueAtom)),
        isDebouncingAtom,
        clearTimeoutAtom,
        debouncedValueAtom,
    };
}
