import { Event } from '@/types/events';
import { UIRequestProto } from '@/types/generated/ui';

export type b6Route = 'startup' | 'stack';

export const fetchB6 = async (
    route: b6Route,
    request: UIRequestProto & { logEvent: Event }
) => {
    return fetch(`/api/${route}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    });
};
