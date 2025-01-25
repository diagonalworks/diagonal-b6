import clsx from "clsx";
import Link from "@docusaurus/Link";
import Heading from "@theme/Heading";
import styles from "./styles.module.css";

type FeatureItem = {
	title: string;
	Svg: React.ComponentType<React.ComponentProps<"svg">>;
	description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
	{
		title: "API",
		// Svg: require('@site/static/img/undraw_docusaurus_mountain.svg').default,
		description: (
			<>The b6 gRPC API; most commonly consumed through the Python library.</>
		),
		link: "/docs/api",
		label: "Documentation",
	},
	{
		title: "Backend",
		description: (
			<>
				The b6 backend, written in Go, providing the webserver, gRPC API server,
				map tiles, custom rendering, and various ingest/post-processing tools.
			</>
		),
		link: "/docs/backend",
		label: "Documentation",
	},
	{
		title: "Frontend",
		description: <>The b6 frontend, written in React, providing the map UI.</>,
		link: "/docs/frontend",
		label: "Documentation",
	},
];

function Feature({ title, Svg, description, link, label }: FeatureItem) {
	// <div className="text--center">
	// 	<Svg className={styles.featureSvg} role="img" />
	// </div>
	return (
		<div className={clsx("col col--4")}>
			<div className="text--center padding-horiz--md">
				<Heading as="h3">{title}</Heading>
				<p>{description}</p>
				<div className={styles.buttons}>
					<Link className="button button--secondary button--lg" to={link}>
						Docs
					</Link>
				</div>
			</div>
		</div>
	);
}

export default function HomepageFeatures(): JSX.Element {
	return (
		<section className={styles.features}>
			<div className="container">
				<div className="row">
					{FeatureList.map((props, idx) => (
						<Feature key={idx} {...props} />
					))}
				</div>
			</div>
		</section>
	);
}
