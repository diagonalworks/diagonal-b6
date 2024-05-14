import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

const INITIAL_LOCATION = {
    mapCenter: {
        lat: 51.5361156,
        lng: -1.255161,
    },
    zoom: 16,
};

export type LocationState = {
    mapCenter: { lat: number; lng: number };
    zoom: number;
    actions: {
        setMapCenter: (lat: number, lng: number) => void;
        setZoom: (zoom: number) => void;
    };
};

// Create the store
const useLocationStore = create<LocationState>()(
    immer((set) => ({
        ...INITIAL_LOCATION,
        actions: {
            setMapCenter: (lat: number, lng: number) =>
                set((state) => ({ ...state, mapCenter: { lat, lng } })),
            setZoom: (zoom: number) => set((state) => ({ ...state, zoom })),
        },
    }))
);

// Update the URL query parameters
const updateUrlParams = (params: URLSearchParams) => {
    const url = new URL(window.location.href);
    url.search = params.toString();
    window.history.replaceState({}, '', url.toString());
};

// Subscribe to changes in the store and update the URL query parameters
useLocationStore.subscribe((state) => {
    const params = new URLSearchParams();
    params.set('ll', `${state.mapCenter.lat},${state.mapCenter.lng}`);
    params.set('z', state.zoom.toString());
    updateUrlParams(params);
});

// Parse the URL query parameters and initialize the store
const initStoreFromUrlParams = () => {
    const params = new URLSearchParams(window.location.search);
    const mapCenter = params.get('ll')?.split(',');
    const zoom = params.get('z');

    if (mapCenter && mapCenter.length === 2) {
        const lat = parseFloat(mapCenter[0]);
        const lng = parseFloat(mapCenter[1]);
        useLocationStore.setState((state) => {
            state.mapCenter = { lat, lng };
        });
    }
    if (zoom) {
        useLocationStore.setState((state) => {
            state.zoom = parseInt(zoom);
        });
    }
};

// Initialize the store from the URL query parameters
initStoreFromUrlParams();

export { useLocationStore };
