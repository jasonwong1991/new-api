import React from 'react';
import { Button } from '@douyinfe/semi-ui';

const InvitationsActions = ({
  setEditingInvitation,
  setShowEdit,
  batchCopyInvitations,
  t,
}) => {
  const handleAddInvitation = () => {
    setEditingInvitation({ id: undefined });
    setShowEdit(true);
  };

  return (
    <div className='flex flex-wrap gap-2 w-full md:w-auto order-2 md:order-1'>
      <Button
        type='primary'
        className='flex-1 md:flex-initial'
        onClick={handleAddInvitation}
        size='small'
      >
        {t('添加邀请码')}
      </Button>

      <Button
        type='tertiary'
        className='flex-1 md:flex-initial'
        onClick={batchCopyInvitations}
        size='small'
      >
        {t('复制所选邀请码')}
      </Button>
    </div>
  );
};

export default InvitationsActions;
