import { useChartDimensions } from '@/lib/useChartDimensions';
import colors from '@/tokens/colors.json';
import { Feature } from '@/types/features';
import { MapView } from '@deck.gl/core/typed';
import { MVTLayer } from '@deck.gl/geo-layers/typed';
import { rgb } from 'd3-color';
import { DeckGL } from 'deck.gl/typed';
import { P, match } from 'ts-pattern';

const InitalZoom = 16;
const InitialCenter = { latE7: 515361156, lngE7: -1255161 };

const hexToRgb = (hex: string | null) => {
    if (!hex) return 'transparent';
    const rgbColor = rgb(hex);
    return [rgbColor.r, rgbColor.g, rgbColor.b];
};

const BACKGROUND_FILL = colors.graphite[20];

export function Map() {
    const [ref, dimensions] = useChartDimensions({});

    const mapView = new MapView({
        id: 'map',
    });

    const layer = new MVTLayer({
        /* This should be fetched from same origin as app, but as the new frontend
        is currently not integrated, we're fetching from diagonal.works for now */
        data: ['https://baseline.diagonal.works/tiles/base/{z}/{x}/{y}.mvt'],
        minZoom: 10,
        maxZoom: 16,
        /* renderSubLayers: (props) => {
            console.log(props);
            return [
                new GeoJ
            ];
        }, */
        getFillColor: (f: Feature) => {
            const color = match(f.properties)
                .with({ layerName: 'background' }, () => BACKGROUND_FILL)
                .with({ layerName: 'water' }, () => colors.blue[20])
                .with({ layerName: 'building' }, () => '#fff')
                .with(
                    { layerName: 'road', highway: P.not(P.nullish) },
                    () => '#e1e1ee'
                )
                .with({ layerName: 'landuse' }, (lf) => {
                    return match(lf)
                        .with(
                            {
                                landuse: P.union(
                                    'park',
                                    'grass',
                                    'pitch',
                                    'park',
                                    'garden',
                                    'playground',
                                    'nature-reserve'
                                ),
                            },
                            () => '#e1e1ee'
                        )
                        .with(
                            {
                                landuse: P.union(
                                    'residential',
                                    'commercial',
                                    'industrial',
                                    'forest'
                                ),
                            },
                            () => '#c5cadd'
                        )
                        .with({ landuse: 'meadow' }, () => '#dbdeeb')
                        .otherwise(() => BACKGROUND_FILL);
                })
                .otherwise(() => BACKGROUND_FILL);

            return hexToRgb(color);
        },
        getLineColor: (f: Feature) => {
            const color = match(f.properties)
                .with(
                    {
                        layerName: 'road',
                        highway: P.not(P.nullish),
                    },
                    () => '#9aa4cc'
                )
                .with(
                    { layerName: 'boundary', waterway: 'coastline' },
                    () => colors.graphite[80]
                )
                .with(
                    { layerName: 'water', waterway: P.not(P.nullish) },
                    () => colors.blue[20]
                )
                .with(
                    {
                        layerName: 'building',
                    },
                    () => '#4f5a7d'
                )
                .otherwise(() => null);
            return hexToRgb(color);
        },
        getLineWidth: (f: Feature) => {
            const width = match(f.properties)
                .with(
                    {
                        layerName: 'building',
                    },
                    () => 0.3
                )
                .with(
                    {
                        layerName: 'road',
                        highway: P.not(P.nullish),
                    },
                    (rf) =>
                        match(rf.highway)
                            .with('motorway', () => 36)
                            .with('trunk', () => 36)
                            .with('primary', () => 32)
                            .with('secondary', () => 26)
                            .with('tertiary', () => 26)
                            .with('street', () => 18)
                            .with('unclassified', () => 18)
                            .with('service', () => 18)
                            .with('residential', () => 18)
                            .with('cycleway', () => 8)
                            .with('footway', () => 8)
                            .with('path', () => 8)
                            .otherwise(() => 0.1)
                )
                .otherwise(() => 0);
            return width / 2;
        },
        lineWidthMinPixels: 0,
        loadOptions: {
            headers: {
                'Access-Control-Allow-Origin': '*',
            },
        },
    });

    return (
        <div
            ref={ref}
            className="w-full h-[80vh] border border-graphite-20 shadow-lg rounded"
        >
            <DeckGL
                style={{
                    width: `${dimensions.width}px`,
                    height: `${dimensions.height}px`,
                    position: 'relative',
                }}
                controller={true}
                views={[mapView]}
                layers={[layer]}
                initialViewState={{
                    longitude: InitialCenter.lngE7 / 1e7,
                    latitude: InitialCenter.latE7 / 1e7,
                    zoom: InitalZoom,
                }}
            />
        </div>
    );
}