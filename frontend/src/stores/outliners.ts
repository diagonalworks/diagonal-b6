import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

import { ImmerStateCreator } from '@/lib/zustand';
import { World } from '@/stores/worlds';
import { UIRequestProto } from '@/types/generated/ui';
import { StackResponse } from '@/types/stack';

/**
 * The specification for an outliner. This is used to store the properties of an outliner and the request needed to fetch its data from the API.
 */
export interface OutlinerSpec {
    id: string;
    world: World['id'];
    properties: {
        active: boolean;
        docked: boolean;
        transient: boolean;
        show: boolean;
        type: 'core' | 'comparison';
        coordinates?: {
            x: number;
            y: number;
        };
    };
    // The request to fetch data for the outliner
    request?: UIRequestProto;
    // Fallback data for outliner, if present, will be used instead of fetching data from the API.
    data?: StackResponse;
}

interface OutlinersStore {
    outliners: Record<string, OutlinerSpec>;
    actions: {
        /**
         * Add an outliner to the store
         * @param spec - The outliner spec to add
         * @returns void
         */
        add: (spec: OutlinerSpec) => void;
        /**
         * Remove an outliner from the store
         * @param id - The id of the outliner to remove
         * @returns void
         */
        remove: (id: string) => void;
        /**
         * Move an outliner by a given amount
         * @param id - The id of the outliner to move
         * @param dx - The amount to move the outliner in the x direction
         * @param dy - The amount to move the outliner in the y direction
         * @returns void
         */
        move: (id: string, dx: number, dy: number) => void;
        /**
         * Set an outliner as active or inactive
         * @param id - The id of the outliner to set
         * @param active - Whether the outliner should be active
         * @returns void
         */
        setActive: (id: string, active: boolean) => void;
        /**
         * Set the API request to fetch data for an outliner
         * @param id - The id of the outliner to set
         * @param request - The request to set
         * @returns void
         */
        setRequest: (id: string, request: UIRequestProto) => void;
        /**
         * Set an outliner as transient or not
         * @param id - The id of the outliner to set
         * @param transient - Whether the outliner should be transient
         * @returns void
         */
        setTransient: (id: string, transient: boolean) => void;
        /**
         * Set the visibility of the layers of an outliner
         * @param id - The id of the outliner to set
         * @param show - Whether the outliner should be visible
         * @returns void
         */
        setVisibility: (id: string, show: boolean) => void;
        /**
         * Get all outliners for a given world
         * @param world - The world to get outliners for
         * @returns An array of outliner specs
         */
        getByWorld: (world: World['id']) => OutlinerSpec[];
    };
}

export const createOutlinersStore: ImmerStateCreator<
    OutlinersStore,
    OutlinersStore
> = (set, get) => ({
    outliners: {},
    actions: {
        add: (spec) => {
            set((state) => {
                for (const id in state.outliners) {
                    if (
                        state.outliners[id].properties.transient &&
                        state.outliners[id].world === spec.world
                    ) {
                        delete state.outliners[id];
                    }
                }
                state.outliners[spec.id] = spec;
            });
        },
        remove: (id) => {
            set((state) => {
                delete state.outliners[id];
            });
        },
        move: (id, dx, dy) => {
            set((state) => {
                const { coordinates } = state.outliners[id].properties;
                if (!coordinates) return;
                state.outliners[id].properties.coordinates = {
                    x: coordinates.x + dx,
                    y: coordinates.y + dy,
                };
            });
        },
        setActive: (id, active) => {
            set((state) => {
                state.outliners[id].properties.active = active;
            });
        },
        setVisibility: (id, show) => {
            set((state) => {
                state.outliners[id].properties.show = show;
            });
        },
        setTransient: (id, transient) => {
            set((state) => {
                state.outliners[id].properties.transient = transient;
            });
        },
        getByWorld: (world) => {
            return Object.values(get().outliners).filter(
                (outliner) => outliner.world === world
            );
        },
        setRequest: (id, request) => {
            set((state) => {
                state.outliners[id].request = request;
            });
        },
    },
});

/**
 * A hook to access the outliners store, which contains all outliners in the app.
 * This is a zustand store that uses immer for immutability.
 * @returns The outliners store
 */
export const useOutlinersStore = create(immer(createOutlinersStore));
