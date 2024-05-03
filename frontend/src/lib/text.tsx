import React from 'react';

export const toTitleCase = (str: string) => {
    return str.replace(/\w\S*/g, function (txt) {
        return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase();
    });
};

export const getNodeText = (node: React.ReactNode): string => {
    if (node == null) return '';

    switch (typeof node) {
        case 'string':
        case 'number':
            return node.toString();

        case 'boolean':
            return '';

        case 'object': {
            if (node instanceof Array) return node.map(getNodeText).join('');

            if ('props' in node) return getNodeText(node.props.children);
            return '';
        }

        default:
            console.warn('Unresolved `node` of type:', typeof node, node);
            return '';
    }
};

export function highlighted(string: string, matches: [number, number][]) {
    const substrings = [];
    let previousEnd = 0;

    for (let [start, end] of matches) {
        const prefix = string.substring(previousEnd, start);
        const match = <strong>{string.substring(start, end)}</strong>;

        substrings.push(prefix, match);
        previousEnd = end;
    }

    substrings.push(string.substring(previousEnd));

    return <span>{React.Children.toArray(substrings)}</span>;
}
