import { Event } from '@/types/events';
import { EvaluateRequestProto } from '@/types/generated/api';
import { UIRequestProto } from '@/types/generated/ui';

export type b6Route = 'startup' | 'stack';

export const b6Path = `${import.meta.env.VITE_B6_BASE_PATH}`;

const stack = async (request: UIRequestProto & { logEvent: Event }) => {
    return fetch(`${b6Path}stack`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    });
};

const startup = async (request: UIRequestProto) => {
    return fetch(`${b6Path}startup`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    });
};

const evaluate = async (request: EvaluateRequestProto) => {
    return fetch(`${b6Path}evaluate`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })
        .then((res) => {
            return res.json();
        })
        .catch((e) => {
            console.error(e);
            return {
                error: e,
            };
        });
};

export const b6 = {
    stack,
    startup,
    evaluate,
};
