import { $IntentionalAny } from '@/utils/defs';

const TOKEN = process.env.FIGMA_TOKEN;

export const figma = () => {
    if (!TOKEN) {
        throw new Error('FIGMA_TOKEN is not defined');
    }

    const headers = {
        'X-Figma-Token': TOKEN,
    };

    const fetchAPI = async (resource: string) => {
        const response = await fetch(`https://api.figma.com/v1${resource}`, {
            headers: {
                Accept: 'application/json',
                ...headers,
            },
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        return response.json();
    };

    const api = {
        file: (file: string) => {
            return fetchAPI(`/files/${file}`);
        },
        nodes: (file: string, nodes: string[]) => {
            return fetchAPI(`/files/${file}/nodes?ids=${nodes.join(',')}`);
        },
        styles: async (file: string) => {
            const response = await fetchAPI(`/files/${file}/styles`);
            const {
                meta: { styles },
            } = response;
            const nodes = styles.map((x: $IntentionalAny) => x.node_id);
            return api.nodes(file, nodes);
        },
    };
    return api;
};
