import { UIRequestProto } from '@/types/generated/ui';

export type b6Event = 's' | 'do' | 'mlc' | 'mfc' | 'oc' | 'os' | 'ws' | 'err';
export type b6Route = 'startup' | 'stack';

export const fetchB6 = async (
    route: b6Route,
    request: UIRequestProto & {
        context?: {
            namespace: string;
            type: string;
        };
    }
) => {
    return fetch(`/api/${route}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    });
};
