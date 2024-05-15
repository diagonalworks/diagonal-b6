import { Event } from '@/types/events';
import { EvaluateRequestProto } from '@/types/generated/api';
import { ComparisonRequestProto, UIRequestProto } from '@/types/generated/ui';
import { StartupResponse } from '@/types/startup';

export type b6Route = 'startup' | 'stack';

export const b6Path = `${import.meta.env.VITE_B6_BASE_PATH}`;

const formatResponse = (res: Response) => {
    if (!res.ok) {
        return res.text().then((v) =>
            Promise.reject({
                status: res.status,
                statusText: res.statusText,
                message: v,
            })
        );
    }
    return res.json();
};

const stack = async (request: UIRequestProto & { logEvent: Event }) => {
    return fetch(`${b6Path}stack`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    }).then((res) => formatResponse(res));
};

const startup = async (urlParams: {
    z: string;
    r: string;
}): Promise<StartupResponse> => {
    return fetch(`${b6Path}startup?` + new URLSearchParams(urlParams)).then(
        (res) => formatResponse(res)
    );
};

const evaluate = async (request: EvaluateRequestProto) => {
    return fetch(`${b6Path}evaluate`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    }).then((res) => {
        return formatResponse(res);
    });
};

const compare = async (request: ComparisonRequestProto) => {
    return fetch(`${b6Path}compare`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    }).then((res) => {
        return formatResponse(res);
    });
};

export const b6 = {
    stack,
    startup,
    evaluate,
    compare,
};
