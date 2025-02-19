CREATE TYPE category_type AS ENUM (
    'Menu',
    'Profile Setting',
    'Setting',
    'Table',
    'Permission',
    'Resource',
    'API keys',
    'Custom endpoint',
    'UserInvite',
    'Fuctions',
    'Activity Logs',
    'Layout',
    'Fields',
    'Field settings',
    'Field type',
    'Relation',
    'Action',
    'Custom error',
    'Calendar view',
    'Microfrontend',
    'LoginPage'
);

ALTER TABLE "language"
ADD COLUMN "category" category_type NOT NULL DEFAULT 'Menu',
ADD COLUMN "platform" VARCHAR(100) DEFAULT 'Admin' NOT NULL;

UPDATE "language" SET category = 'Menu' WHERE key IN ('Settings', 'Users', 'Create', 'Files', 'Create folder', 'Media', 'Show', 'Users', 'Add table', 'Create table', 'Attach to table', 'Attach to microfrontend', 'Add params', 'Website Link', 'Save', 'Add microfrontend', 'Add website', 'Add folder');
UPDATE "language" SET category = 'Profile Setting' WHERE key IN ('Languages', 'Logout', 'Log out of your account', 'You will need to log back in to access your workspace.', 'Log out', 'Cancel');
UPDATE "language" SET category = 'Setting' WHERE key IN ('Project Settings', 'Permissions', 'Resources', 'Code', 'Activity Logs', 'Environments', 'Versions', 'Language Control', 'Functions', 'Microfrontend', 'Upload ERD', 'Data', 'Models', 'Name', 'Language', 'Currency', 'Timezone', 'Description', 'Add', 'Self recover', 'Self register', 'Table', 'Session limit', 'Matrix Details', 'Role', 'Connection', 'View slug', 'Table slug', 'Resource settings', 'API keys', 'Create new');
UPDATE "language" SET category = 'Table' WHERE key IN ('Relations', 'Actions', 'Custom errors', 'View fields', 'Disable Edit table', 'Enable Multi language', 'Additional', 'Default editable', 'Add fields', 'Seaarch by filled name', 'Fields', 'Details', 'Search', 'Title', 'Default page link', 'Items', 'Select All', 'Name', 'General', 'Key', 'Login table', 'Cache', 'Soft delete', 'Sort', 'Layout', 'Columns', 'Group', 'Tab group', 'Fix column', 'Export', 'Import', 'Delete', 'Docs', 'Visible columns', 'Search by filled name', 'Show all', 'Group columns', 'Tab group columns', 'Fix columns', 'Upload file', 'Confirmation', 'Drag and drop files here', 'Edit field', 'Sort A -> Z', 'Add Summary', 'Fix column', 'Hide field', 'Delete field', 'View', 'Create item');
UPDATE "language" SET category = 'Permission' WHERE key IN ('Creatable', 'Relation Buttons', 'Create relation', 'Table', 'Menu', 'Global Permission', 'Objects', 'Record Permission', 'Field Permission', 'Action Permission', 'Relation Permission', 'View Permission', 'Custom Permission', 'Reading', 'Adding', 'Editing', 'Deleting', 'Public');
UPDATE "language" SET category = 'Resource' WHERE key IN ('Resource Info', 'Type');
UPDATE "language" SET category = 'API keys' WHERE key IN ('Client ID', 'Platform name', 'Monthly limit', 'RPS limit', 'Used count');
UPDATE "language" SET category = 'Custom endpoint' WHERE key IN ('Redirects', 'From', 'Created at', 'Updated at');
UPDATE "language" SET category = 'UserInvite' WHERE key IN ('Mail', 'Phone', 'Invite', 'Invite user', 'Password', 'User type', 'More', 'Email', 'Active role');
UPDATE "language" SET category = 'Fuctions' WHERE key IN ('Faas functions', 'Status', 'Path');
UPDATE "language" SET category = 'Activity Logs' WHERE key IN ('Action', 'Collection', 'Action On', 'Action By');
UPDATE "language" SET category = 'Layout' WHERE key IN ('Default', 'Modal', 'Remove tabs', 'Form fields', 'Relation tabs', 'Section tabs', 'Relation table', 'Add section tab', 'Add section', 'View options');
UPDATE "language" SET category = 'Fields' WHERE key IN ('Field Label', 'Field Type', 'Field Slug', 'Field create');
UPDATE "language" SET category = 'Field settings' WHERE key IN ('Default value', 'Schema', 'Validation', 'Autofill', 'Auto filter', 'Field hide', 'Error message', 'Disabled', 'Required', 'Duplicate', 'Autofill table', 'Autofilter field', 'Automatic', 'Hide field from');
UPDATE "language" SET category = 'Field type' WHERE key IN ('Single Line', 'Multi Line', 'Date', 'Date time', 'Date time - timezone', 'Time', 'Number', 'Float', 'Money', 'Checkbox', 'Switch', 'Select', 'Status', 'Point', 'Geozone', 'Photo', 'Multi Image', 'Video', 'File', 'Formula', 'Formula in frontend', 'Formula in backend', 'Text', 'Link', 'Person', 'Button', 'Incremenet ID', 'Internation phone', 'Email', 'Icon', 'Password', 'Color');
UPDATE "language" SET category = 'Relation' WHERE key IN ('Table from', 'Table to', 'Relation type');
UPDATE "language" SET category = 'Action' WHERE key IN ('Action type', 'Redirect URL', 'Method', 'No limit');
UPDATE "language" SET category = 'Custom error' WHERE key IN ('Message', 'Fields list', 'Error id');
UPDATE "language" SET category = 'Calendar view' WHERE key IN ('Browse', 'Today', 'Day', 'Calendar settings', 'Time from', 'Time to');
UPDATE "language" SET category = 'Microfrontend' WHERE key IN ('Framework type');
UPDATE "language" SET category = 'LoginPage' WHERE key IN ('Login');
