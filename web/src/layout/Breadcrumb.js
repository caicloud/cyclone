import React from 'react';
import { Link } from 'react-router-dom';
import { Breadcrumb } from 'antd';
import PropTypes from 'prop-types';

const mainModules = [
  'overview',
  'project',
  'stageTemplate',
  'resource',
  'workflow',
];
const operations = ['update', 'add'];

/**
 * define the route rules
 * list page => /project
 * project detail  => /project/:projectId
 * update project => /project/:projectId/update
 * add project => /project/add
 */
const BreadcrumbComponent = ({ location }) => {
  const pathSnippets = location.pathname.split('/').filter(i => i);
  const extraBreadcrumbItems = pathSnippets.map((path, index) => {
    const url = `/${pathSnippets.slice(0, index + 1).join('/')}`;
    let text = path;
    if (_.includes(mainModules, path)) {
      text = intl.get(`sideNav.${path}`);
    } else if (_.includes(operations, path)) {
      text = intl.get(`operation.${path}`);
    }
    return (
      <Breadcrumb.Item key={url}>
        <Link to={url}>{text}</Link>
      </Breadcrumb.Item>
    );
  });
  return (
    <Breadcrumb style={{ marginBottom: '12px' }}>
      {extraBreadcrumbItems}
    </Breadcrumb>
  );
};

BreadcrumbComponent.propTypes = {
  location: PropTypes.object,
};

export default BreadcrumbComponent;
