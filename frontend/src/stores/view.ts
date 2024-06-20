import { ViewState } from 'react-map-gl';
import { StateCreator, create } from 'zustand';

import { usePersistURL } from '@/hooks/usePersistURL';

/**
 * Interface representing the view store. The view store contains the map center and zoom level.
 */
interface ViewStore {
    view: Partial<ViewState>;
    initialView: Partial<ViewState>;
    actions: {
        /**
         * Sets the view state.
         * @param view - The partial view state to set.
         */
        setView: (view: Partial<ViewState>) => void;
    };
}

export const createViewStore: StateCreator<ViewStore> = (set) => ({
    view: {},
    initialView: {},
    actions: {
        setView: (view) => set({ view }),
    },
});

/**
 * Hook for using the view store. This store contains the view state and an action to set the view state.
 * @returns The view store.
 */
export const useViewStore = create(createViewStore);

/**
 * Type representing the URL parameters for the view.
 */
type ViewURLParams = {
    ll?: string;
    z?: string;
};

const encode = (state: Partial<ViewStore>): ViewURLParams => ({
    ll:
        state.view?.latitude && state.view?.longitude
            ? `${state.view.latitude},${state.view.longitude}`
            : '',
    z: state.view?.zoom ? state.view.zoom.toString() : '',
});

const decode = (
    params: ViewURLParams,
    initial?: boolean
): ((state: ViewStore) => ViewStore) => {
    return (state) => ({
        ...state,
        ...(initial && {
            initialView: {
                latitude: params.ll
                    ? parseFloat(params.ll.split(',')[0])
                    : undefined,
                longitude: params.ll
                    ? parseFloat(params.ll.split(',')[1])
                    : undefined,
                zoom: params.z ? parseInt(params.z) : undefined,
            },
        }),
        view: {
            ...state.view,
            latitude: params.ll
                ? parseFloat(params.ll.split(',')[0])
                : state.view?.latitude,
            longitude: params.ll
                ? parseFloat(params.ll.split(',')[1])
                : state.view?.longitude,
            zoom: params.z ? parseInt(params.z) : state.view?.zoom,
        },
    });
};

/**
 * Hook for using URL persistence for the view store.
 * @returns The view store with URL persistence.
 */
export const useViewURLStorage = () => {
    return usePersistURL(useViewStore, encode, decode);
};
