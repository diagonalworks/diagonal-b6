import { Event } from '@/types/events';
import { UIRequestProto } from '@/types/generated/ui';

export type b6Route = 'startup' | 'stack';

export const b6Path = `${import.meta.env.VITE_B6_BASE_PATH}`;

export const fetchB6 = async (
    route: b6Route,
    request: UIRequestProto & { logEvent: Event }
) => {
    return fetch(`${b6Path}${route}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    });
};
