import React from 'react';
import { Link } from 'react-router-dom';
import { Breadcrumb } from 'antd';
import PropTypes from 'prop-types';

const mainModules = [
  'overview',
  'projects',
  'stageTemplate',
  'resource',
  'integration',
];
const operations = ['update', 'add'];

/**
 * define the route rules
 * list page => /projects
 * project detail  => /projects/:projectId
 * update project => /projects/:projectId/update
 * add project => /projects/add
 */
const BreadcrumbComponent = ({ location }) => {
  const pathSnippets = location.pathname.split('/').filter(i => i);
  const extraBreadcrumbItems = pathSnippets.map((path, index) => {
    const url = `/${pathSnippets.slice(0, index + 1).join('/')}`;
    let text = path;
    if (_.includes(mainModules, path)) {
      text = <Link to={url}>{intl.get(`sideNav.${path}`)}</Link>;
    } else if (_.includes(operations, path)) {
      text = intl.get(`operation.${path}`);
    } else if (_.includes(mainModules, pathSnippets[index - 1])) {
      text = <Link to={url}>{path}</Link>;
    }
    return <Breadcrumb.Item key={url}>{text}</Breadcrumb.Item>;
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
