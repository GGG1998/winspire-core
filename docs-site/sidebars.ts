import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  // Main domains sidebar
  domainsSidebar: [
    {
      type: 'doc',
      id: 'README',
      label: 'Overview',
    },
    {
      type: 'category',
      label: 'Core Domain',
      collapsed: false,
      items: [
        'core/README',
        {
          type: 'category',
          label: 'Users Subdomain',
          items: [
            'core/users/README',
            'core/users/data-model',
            'core/users/policies',
            'core/users/integration',
            {
              type: 'category',
              label: 'API',
              items: [
                {
                  type: 'link',
                  label: 'REST API (OpenAPI)',
                  href: '/core/users/api/openapi.yaml',
                },
              ],
            },
          ],
        },
      ],
    },
  ],

  // Platform sidebar
  platformSidebar: [
    {
      type: 'doc',
      id: 'platform/README',
      label: 'Platform Overview',
    },
    {
      type: 'doc',
      id: 'platform/architecture-overview',
      label: 'Architecture',
    },
    {
      type: 'doc',
      id: 'platform/api-standards',
      label: 'API Standards',
    },
    {
      type: 'doc',
      id: 'platform/integration-guidelines',
      label: 'Integration Guidelines',
    },
    {
      type: 'category',
      label: 'Shared Schemas',
      collapsed: false,
      items: [
        {
          type: 'link',
          label: 'User Common',
          href: '/platform/shared-schemas/user-common.yaml',
        },
        {
          type: 'link',
          label: 'Error Codes',
          href: '/platform/shared-schemas/error-codes.yaml',
        },
      ],
    },
  ],

  // Reference sidebar
  referenceSidebar: [
    {
      type: 'doc',
      id: 'reference/README',
      label: 'Reference Overview',
    },
    {
      type: 'doc',
      id: 'reference/glossary',
      label: 'Glossary',
    },
    {
      type: 'doc',
      id: 'reference/contributing',
      label: 'Contributing',
    },
    {
      type: 'category',
      label: 'Architecture Decision Records',
      collapsed: false,
      items: [
        'reference/adr/template',
        'reference/adr/domain-driven-architecture',
      ],
    },
  ],
};

export default sidebars;
