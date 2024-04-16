import { urlSearchParamsStorage } from '@/lib/storage';
import { atom } from 'jotai';
import { atomWithStorage } from 'jotai/utils';
import { ViewState } from 'react-map-gl';

const INITIAL_COORDINATES = { latE7: 515361156, lngE7: -1255161 };

export const zoomAtom = atomWithStorage<number>(
    'z',
    16,
    urlSearchParamsStorage({}),
    {
        getOnInit: true,
    }
);

export const centerAtom = atomWithStorage<{ lat: number; lng: number }>(
    'll',
    {
        lat: INITIAL_COORDINATES.latE7 / 1e7,
        lng: INITIAL_COORDINATES.lngE7 / 1e7,
    },
    urlSearchParamsStorage({
        serialize: (value) =>
            value.lat && value.lng ? `${value.lat},${value.lng}` : '',
        deserialize: (value) => {
            if (!value)
                return {
                    lat: INITIAL_COORDINATES.latE7 / 1e7,
                    lng: INITIAL_COORDINATES.lngE7 / 1e7,
                };
            const [lat, lng] = value.split(',').map(Number);
            return { lat, lng };
        },
    }),
    {
        getOnInit: true,
    }
);

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
        set(centerAtom, { lat: view.latitude, lng: view.longitude });
        set(zoomAtom, view.zoom);
    }
);
