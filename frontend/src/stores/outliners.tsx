import { ImmerStateCreator } from '@/lib/zustand';
import { UIRequestProto } from '@/types/generated/ui';
import { StackResponse } from '@/types/stack';
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { World } from './worlds';

export interface OutlinerSpec {
    id: string;
    world: World['id'];
    properties: {
        active: boolean;
        docked: boolean;
        transient: boolean;
        type: 'core' | 'comparison';
        coordinates?: {
            x: number;
            y: number;
        };
    };
    request?: UIRequestProto;
    data?: StackResponse;
}

interface OutlinersStore {
    outliners: Record<string, OutlinerSpec>;
    actions: {
        add: (spec: OutlinerSpec) => void;
        remove: (id: string) => void;
        move: (id: string, dx: number, dy: number) => void;
        setActive: (id: string, active: boolean) => void;
        setDocked: (id: string, docked: boolean) => void;
        setRequest: (id: string, request: UIRequestProto) => void;
        setTransient: (id: string, transient: boolean) => void;
        getByWorld: (world: World['id']) => OutlinerSpec[];
        getById: (id: string) => OutlinerSpec;
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
        setDocked: (id, docked) => {
            set((state) => {
                state.outliners[id].properties.docked = docked;
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
        getById: (id) => {
            return get().outliners[id];
        },
        setRequest: (id, request) => {
            set((state) => {
                state.outliners[id].request = request;
            });
        },
    },
});

export const useOutlinersStore = create(immer(createOutlinersStore));
