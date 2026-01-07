import React, { useMemo, useState } from 'react';
import { Empty } from '@douyinfe/semi-ui';
import CardTable from '../../common/ui/CardTable';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { getInvitationsColumns } from './InvitationsColumnDefs';
import DeleteInvitationModal from './modals/DeleteInvitationModal';

const InvitationsTable = (invitationsData) => {
  const {
    invitations,
    loading,
    activePage,
    pageSize,
    totalCount,
    compactMode,
    handlePageChange,
    handlePageSizeChange,
    rowSelection,
    handleRow,
    manageInvitation,
    copyText,
    setEditingInvitation,
    setShowEdit,
    refresh,
    t,
  } = invitationsData;

  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [deletingRecord, setDeletingRecord] = useState(null);

  const showDeleteInvitationModal = (record) => {
    setDeletingRecord(record);
    setShowDeleteModal(true);
  };

  const columns = useMemo(() => {
    return getInvitationsColumns({
      t,
      manageInvitation,
      copyText,
      setEditingInvitation,
      setShowEdit,
      showDeleteInvitationModal,
    });
  }, [t, manageInvitation, copyText, setEditingInvitation, setShowEdit]);

  const tableColumns = useMemo(() => {
    return compactMode
      ? columns.map((col) => {
          if (col.dataIndex === 'operate') {
            const { fixed, ...rest } = col;
            return rest;
          }
          return col;
        })
      : columns;
  }, [compactMode, columns]);

  return (
    <>
      <CardTable
        columns={tableColumns}
        dataSource={invitations}
        scroll={compactMode ? undefined : { x: 'max-content' }}
        pagination={{
          currentPage: activePage,
          pageSize: pageSize,
          total: totalCount,
          showSizeChanger: true,
          pageSizeOptions: [10, 20, 50, 100],
          onPageSizeChange: handlePageSizeChange,
          onPageChange: handlePageChange,
        }}
        hidePagination={true}
        loading={loading}
        rowSelection={rowSelection}
        onRow={handleRow}
        empty={
          <Empty
            image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
            darkModeImage={
              <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
            }
            description={t('搜索无结果')}
            style={{ padding: 30 }}
          />
        }
        className='rounded-xl overflow-hidden'
        size='middle'
      />

      <DeleteInvitationModal
        visible={showDeleteModal}
        onCancel={() => setShowDeleteModal(false)}
        record={deletingRecord}
        manageInvitation={manageInvitation}
        refresh={refresh}
        t={t}
      />
    </>
  );
};

export default InvitationsTable;
