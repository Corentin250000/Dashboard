**Outils pour un dashboard web**

Frontend

| Nom | Point fort | Point faible |
| :---- | :---- | :---- |
| React | Écosystème immense, très nombreuses librairies de graphiques, communauté très active, flexible | Nécessite d'assembler soi-même beaucoup de briques (routing, state management…), courbe d'apprentissage JSX/hooks |
| Vue.js | Syntaxe simple et progressive, bonne documentation, adapté aux dashboards moyens | Écosystème plus petit, moins de ressources/offres d'emploi que React, librairies tierces parfois moins matures |
| Angular | Framework complet "batteries included" (routing, forms, HTTP intégrés), architecture stricte, TypeScript natif | Courbe d'apprentissage élevée, verbeux, plus lourd pour de petits projets |
| Svelte / SvelteKit | Compile en JS vanilla (pas de virtual DOM), performances excellentes, code très concis | Écosystème plus jeune et restreint, moins de composants dashboard prêts à l'emploi |
| Solid.js | Très performant (proche du JS natif), API proche de React mais sans virtual DOM | Communauté encore petite, peu de librairies dashboard dédiées |
| Chart.js | Simple à prendre en main | Options de personnalisation limitées sur les cas complexes |
| D3.js | Contrôle total, visualisations très avancées | Complexité et temps de développement élevés |
| ApexCharts | Bon rendu visuel out-of-the-box, interactif | Moins flexible que D3 |
| Recharts/Nivo (React) | Bonne intégration React, déclaratif | Performances en baisse sur très gros volumes de données |
| Material UI (MUI) | Très complet, design cohérent | Bundle assez lourd, look "Google" par défaut |
| Ant Design | Nombreux composants orientés admin/dashboard | Personnalisation parfois rigide |
| Tailwind CSS | Flexibilité totale, pas de style imposé | Demande plus de temps de conception, pas de composants "prêts" |

Backend

| Nom | Point fort | Point faible |
| :---- | :---- | :---- |
| Node.js \+ Express | JS partout (front/back), léger, immense écosystème npm, bon pour API REST et temps réel | Peu de structure imposée (risque de code désorganisé), gestion asynchrone parfois piégeuse |
| NestJS | Architecture modulaire type Angular, TypeScript natif, bon pour projets d'équipe/entreprise | Plus verbeux, courbe d'apprentissage plus élevée qu'Express seul |
| Django (Python) | "Batteries included" (admin, ORM, auth), très rapide pour prototyper, sécurité robuste par défaut | Moins flexible que des microframeworks, peut être lourd pour une simple API |
| FastAPI (Python) | Très performant, typage et validation automatique, documentation API autogénérée (Swagger) | Écosystème plus jeune que Django, moins de fonctionnalités "out of the box" |
| Flask (Python) | Minimaliste, flexible, facile à prendre en main | Tout est à ajouter manuellement (auth, ORM…), moins structuré pour gros projets |
| Ruby on Rails | Convention over configuration, très rapide pour du CRUD, mature | Communauté et popularité en déclin, performances inférieures à Node/Go |
| Spring Boot (Java) | Très robuste, adapté aux systèmes critiques/entreprise, écosystème énorme, bon typage | Verbeux, temps de démarrage/compilation plus long, courbe d'apprentissage élevée |
| .NET (C\#) | Performances élevées, excellent tooling, bon support entreprise/Azure | Historiquement associé à l'écosystème Microsoft, communauté open-source plus restreinte |
| Laravel (PHP) | Très complet, ORM (Eloquent) agréable, bonne documentation, rapide à développer | Performances en retrait par rapport à Node/Go, PHP moins prisé pour du greenfield |
| Go (Gin/Fiber) | Performances et concurrence excellentes, binaires légers, bon pour APIs à fort trafic | Écosystème moins riche, plus de code bas niveau, courbe d'apprentissage pour non-initiés |
| Firebase (BaaS) | Mise en place ultra-rapide, temps réel intégré | Vendor lock-in, coûts qui grimpent à l'échelle |
| Supabase (BaaS) | Équivalent open source de Firebase basé sur Postgres | Écosystème plus jeune, moins de fonctionnalités managées |

## Choix retenus pour ce projet

Le projet utilise finalement :

| Besoin | Choix | Justification |
|---|---|---|
| Application web | React + TypeScript | Composants réutilisables, typage statique et bonne intégration avec l'interface de dashboard. |
| Build frontend | Vite | Démarrage rapide, HMR et build de production simple. |
| Mise en page | `react-grid-layout` | Grille 6x6, tailles fixes, déplacement desktop et gestion responsive. |
| Style | Tailwind CSS | Tokens de thème centralisés et itération rapide sur l'interface. |
| API | Go + Fiber | Binaire léger, concurrence native et API HTTP performante. |
| Accès aux données | PostgreSQL + GORM | Relations, JSONB pour les configurations et intégration Go simple. |
| Déploiement local | Docker Compose | Démarrage reproductible du backend, de la base et du frontend. |
| Secrets | Doppler avec repli `.env` | Injection des secrets en environnement Docker tout en conservant un mode local sans compte Doppler. |

Les bibliothèques listées dans les tableaux ci-dessus sont des alternatives
étudiées, pas nécessairement des dépendances du dépôt. Les dépendances
effectivement utilisées sont décrites dans `frontend/package.json` et
`backend/go.mod`.

## Limites actuelles

- Le rafraîchissement des widgets est piloté par le frontend, sans worker
  backend générique.
- La grille mobile et tablette est dérivée de la disposition desktop et n'est
  pas persistée séparément.
- Les fonctionnalités d'administration ne sont pas implémentées.
- Les intégrations OAuth dépendent des quotas, scopes et restrictions de chaque
  fournisseur.
