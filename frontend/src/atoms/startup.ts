import { b6Path } from '@/lib/b6';
import { StartupResponse } from '@/types/startup';
import { atomWithQuery } from 'jotai-tanstack-query';
import { collectionAtom } from './app';
import { viewAtom } from './location';

export const startupQueryAtom = atomWithQuery((get) => {
    const collection = get(collectionAtom);
    const viewState = get(viewAtom);
    return {
        queryKey: ['startup', collection],
        queryFn: () =>
            fetch(
                `${b6Path}startup?` +
                    new URLSearchParams({
                        z: viewState.zoom.toString(),
                        ll: `${viewState.latitude},${viewState.longitude}`,
                        r: collection,
                    })
            ).then((res) => res.json() as Promise<StartupResponse>),
    };
});
