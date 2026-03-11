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