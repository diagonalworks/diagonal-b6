import { $IntentionalAny } from '@/utils/defs';
import { ArgumentParser } from 'argparse';
import { rgb } from 'd3-color';
import fs from 'fs';
import { groupBy, mapValues } from 'lodash';
import { figma } from './api';

const main = async () => {
    const parser = new ArgumentParser({
        description: 'Download color styles from Figma',
    });
    parser.add_argument('-f', '--file', {
        help: 'Figma file id',
        required: true,
    });
    parser.add_argument('-o', '--output', {
        help: 'Output file',
        required: true,
    });

    const { file, output } = parser.parse_args();
    const api = figma();

    const styles = await api.styles(file);

    const colorsFlat = Object.values(styles.nodes).map((n: $IntentionalAny) => {
        const { document } = n;
        return {
            id: document.name,
            name: document.name.split('/')[0].toLowerCase(),
            shade: document.name.split('/')[1],
            color: rgb(
                document.fills[0].color.r * 255,
                document.fills[0].color.g * 255,
                document.fills[0].color.b * 255,
                document.fills[0].color.a,
            ).formatHex(),
        };
    });

    const colors = mapValues(
        groupBy(colorsFlat, (c) => c.name),
        (d) => {
            return d.reduce(
                (acc, curr) => ({ ...acc, [curr.shade]: curr.color }),
                {} as Record<string, string>,
            );
        },
    );

    fs.writeFileSync(output, JSON.stringify(colors, null, 2));
};

main().catch((error) => {
    console.error(error);
    process.exit(1);
});
