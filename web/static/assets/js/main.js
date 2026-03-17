const menuToggle = document.getElementById('menuToggle');
const mobileMenu = document.getElementById('mobileMenu');
const iconOpen   = document.getElementById('iconOpen');
const iconClose  = document.getElementById('iconClose');

menuToggle.addEventListener('click', () => {
    const isOpen = mobileMenu.style.display === 'flex';

    mobileMenu.style.display = isOpen ? 'none' : 'flex';
    mobileMenu.style.flexDirection = 'column';

    iconOpen.classList.toggle('hidden', !isOpen);
    iconClose.classList.toggle('hidden', isOpen);
});

async function triggerGCJob() {
    const confirmed = await Swal.fire({
        title: 'Run Garbage Collection?',
        text: 'This will scan and remove duplicate/near-duplicate facts and orphan media objects.',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#4f46e5',
        cancelButtonColor: '#6b7280',
        confirmButtonText: 'Yes, run it!',
        cancelButtonText: 'Cancel',
    })

    if (!confirmed.isConfirmed) return

    // Loading state
    const btn = document.getElementById('gcTriggerBtn')
    if (btn) btn.disabled = true

    Swal.fire({
        title: 'Starting GC Job...',
        text: 'Please wait.',
        allowOutsideClick: false,
        allowEscapeKey: false,
        didOpen: () => Swal.showLoading(),
    })

    try {
        const res = await fetch('/admin/api/gc/execute', { method: 'POST' })
        const data = await res.json()

        if (!res.ok || data.error) {
            await Swal.fire({
                icon: 'error',
                title: 'Failed to Start GC Job',
                text: data.error ?? 'An unexpected error occurred.',
                confirmButtonColor: '#4f46e5',
            })
            return
        }

        await Swal.fire({
            icon: 'success',
            title: 'GC Job Started!',
            html: `Job ID: <code class="text-indigo-600 font-mono text-sm">${data.job_id}</code>`,
            confirmButtonColor: '#4f46e5',
            confirmButtonText: 'View Job',
        })

        window.location.href = `/admin/gc/jobs/${data.job_id}`

    } catch (err) {
        await Swal.fire({
            icon: 'error',
            title: 'Network Error',
            text: 'Could not reach the server. Please try again.',
            confirmButtonColor: '#4f46e5',
        })
    } finally {
        if (btn) btn.disabled = false
    }
}

async function deleteFact(id) {
    const result = await Swal.fire({
        title: 'Are you sure?',
        text: "You won't be able to revert this!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#4f46e5',
        cancelButtonColor: '#9ca3af',
        confirmButtonText: 'Yes, delete it!'
    });

    if (result.isConfirmed) {
        try {
            const response = await fetch(`/admin/api/facts/${id}`, { method: 'DELETE' });

            if (response.ok) {
                await Swal.fire({
                    title: 'Deleted!',
                    text: `Fun Fact #${id} has been deleted.`,
                    icon: 'success',
                    timer: 1500,
                    showConfirmButton: false
                });
                document.getElementById(`fact-${id}`).remove();

                // window.location.reload();
            } else {
                Swal.fire('Error!', 'Failed to delete the data.', 'error');
            }
        } catch (err) {
            Swal.fire('Error!', 'Something went wrong on the server.', 'error');
            console.error('Error:', err);
        }
    }
}

async function deleteTag(id)  {
    const result = await Swal.fire({
        title: 'Are you sure?',
        text: "You won't be able to revert this!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#4f46e5',
        cancelButtonColor: '#9ca3af',
        confirmButtonText: 'Yes, delete it!'
    });

    if (result.isConfirmed) {
        try {
            const response = await fetch(`/admin/api/tags/${id}`, { method: 'DELETE' });

            if (response.ok) {
                await Swal.fire({
                    title: 'Deleted!',
                    text: `Tag ID #${id} has been deleted.`,
                    icon: 'success',
                    timer: 1500,
                    showConfirmButton: false
                });
                document.getElementById(`tag-${id}`).remove();

                // window.location.reload();
            } else {
                Swal.fire('Error!', 'Failed to delete the data.', 'error');
            }
        } catch (err) {
            Swal.fire('Error!', 'Something went wrong on the server.', 'error');
            console.error('Error:', err);
        }
    }
}

// Scroll progress bar
window.addEventListener('scroll', () => {
    const scrolled = (window.scrollY / (document.body.scrollHeight - window.innerHeight)) * 100;
    document.getElementById('scrollBar').style.width = scrolled + '%';
});

function truncateChars(str, n) {
    if (!str) return ''
    return str.length > n ? str.slice(0, n - 1) + '…' : str
}