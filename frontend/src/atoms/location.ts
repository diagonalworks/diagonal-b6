import { atom } from 'jotai';
import { atomWithHash } from 'jotai-location';
import { ViewState } from 'react-map-gl';

const INITIAL_COORDINATES = { latE7: 515361156, lngE7: -1255161 };

export const zoomAtom = atomWithHash<number>('z', 16);
export const centerAtom = atomWithHash<{ lat: number; lng: number }>(
    'll',
    {
        lat: INITIAL_COORDINATES.latE7 / 1e7,
        lng: INITIAL_COORDINATES.lngE7 / 1e7,
    },
    {
        serialize: (value) => `${value.lat},${value.lng}`,
        deserialize: (value) => {
            const [lat, lng] = value.split(',').map(Number);
            return { lat, lng };
        },
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
