import React from 'react';
import CardPro from '../../common/ui/CardPro';
import InvitationsTable from './InvitationsTable';
import InvitationsActions from './InvitationsActions';
import InvitationsFilters from './InvitationsFilters';
import InvitationsDescription from './InvitationsDescription';
import EditInvitationModal from './modals/EditInvitationModal';
import { useInvitationsData } from '../../../hooks/invitations/useInvitationsData';
import { useIsMobile } from '../../../hooks/common/useIsMobile';
import { createCardProPagination } from '../../../helpers/utils';

const InvitationsPage = () => {
  const invitationsData = useInvitationsData();
  const isMobile = useIsMobile();

  const {
    showEdit,
    editingInvitation,
    closeEdit,
    refresh,
    selectedKeys,
    setEditingInvitation,
    setShowEdit,
    batchCopyInvitations,
    formInitValues,
    setFormApi,
    searchInvitations,
    loading,
    searching,
    compactMode,
    setCompactMode,
    t,
  } = invitationsData;

  return (
    <>
      <EditInvitationModal
        refresh={refresh}
        editingInvitation={editingInvitation}
        visible={showEdit}
        handleClose={closeEdit}
      />

      <CardPro
        type='type1'
        descriptionArea={
          <InvitationsDescription
            compactMode={compactMode}
            setCompactMode={setCompactMode}
            t={t}
          />
        }
        actionsArea={
          <div className='flex flex-col md:flex-row justify-between items-center gap-2 w-full'>
            <InvitationsActions
              selectedKeys={selectedKeys}
              setEditingInvitation={setEditingInvitation}
              setShowEdit={setShowEdit}
              batchCopyInvitations={batchCopyInvitations}
              t={t}
            />

            <div className='w-full md:w-full lg:w-auto order-1 md:order-2'>
              <InvitationsFilters
                formInitValues={formInitValues}
                setFormApi={setFormApi}
                searchInvitations={searchInvitations}
                loading={loading}
                searching={searching}
                t={t}
              />
            </div>
          </div>
        }
        paginationArea={createCardProPagination({
          currentPage: invitationsData.activePage,
          pageSize: invitationsData.pageSize,
          total: invitationsData.totalCount,
          onPageChange: invitationsData.handlePageChange,
          onPageSizeChange: invitationsData.handlePageSizeChange,
          isMobile: isMobile,
          t: invitationsData.t,
        })}
        t={invitationsData.t}
      >
        <InvitationsTable {...invitationsData} />
      </CardPro>
    </>
  );
};

export default InvitationsPage;
