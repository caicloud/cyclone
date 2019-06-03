import { Modal } from 'antd';
import { inject, observer } from 'mobx-react';
import EllipsisMenu from '@/components/public/ellipsisMenu';
import PropTypes from 'prop-types';

const confirm = Modal.confirm;

@inject('project')
@observer
class MenuAction extends React.Component {
  static propTypes = {
    project: PropTypes.object,
    name: PropTypes.string,
    history: PropTypes.object,
    detail: PropTypes.bool,
  };

  removeProject = name => {
    const { project, history, detail } = this.props;
    confirm({
      title: `Do you Want to delete project ${name} ?`,
      onOk() {
        project.deleteProject(name, () => {
          history.replace(`/projects`);
          if (!detail) {
            project.listProjects({
              sort: true,
              ascending: false,
              detail: true,
            });
          }
        });
      },
    });
  };

  updateProject = name => {
    const { history } = this.props;
    history.push(`/projects/${name}/update`);
  };

  render() {
    const { name } = this.props;
    return (
      <EllipsisMenu
        menuText={[intl.get('operation.modify'), intl.get('operation.delete')]}
        menuFunc={[
          () => {
            this.updateProject(name);
          },
          () => {
            this.removeProject(name);
          },
        ]}
      />
    );
  }
}
export default MenuAction;
