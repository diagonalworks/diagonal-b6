import { usePersistURL } from '@/hooks/usePersistURL';
import { ViewState } from 'react-map-gl';
import { create, StateCreator } from 'zustand';

interface ViewStore {
    view: Partial<ViewState>;
    initialView: Partial<ViewState>;
    actions: {
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

export const useViewStore = create(createViewStore);

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

export const useViewURLStorage = () => {
    return usePersistURL(useViewStore, encode, decode);
};
