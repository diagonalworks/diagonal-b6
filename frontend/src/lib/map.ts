import { Feature, GeoJsonProperties, Geometry } from 'geojson';

export const isSamePositionPoints = (
    f1: Feature<Geometry, GeoJsonProperties>,
    f2: Feature<Geometry, GeoJsonProperties>,
    precision: number = 6
) => {
    if (f1.geometry.type !== 'Point' || f2.geometry.type !== 'Point')
        return false;
    return (
        f1.geometry.coordinates[0].toFixed(precision) ===
            f2.geometry.coordinates[0].toFixed(precision) &&
        f1.geometry.coordinates[1].toFixed(precision) ===
            f2.geometry.coordinates[1].toFixed(precision)
    );
};
