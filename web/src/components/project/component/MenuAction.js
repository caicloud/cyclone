import { Modal } from 'antd';
import { inject, observer, PropTypes } from 'mobx-react';
import EllipsisMenu from '../../public/ellipsisMenu';

const confirm = Modal.confirm;

@inject('project')
@observer
class MenuAction extends React.Component {
  static propTypes = {
    project: PropTypes.observableObject,
    history: PropTypes.object,
    name: PropTypes.string,
    detail: PropTypes.bool,
  };
  removeProject = name => {
    const { project, history, detail } = this.props;
    confirm({
      title: `Do you Want to delete project ${name} ?`,
      onOk() {
        project.deleteProject(name, () => {
          history.replace(`/project`);
          if (!detail) {
            project.listProjects();
          }
        });
      },
    });
  };

  updateProject = name => {
    const { history } = this.props;
    history.push(`/project/${name}/update`);
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
