import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/core/constants/app_radius.dart';
import 'package:phone/core/widgets/app_bottom_sheet.dart';
import 'package:phone/core/widgets/app_snack_bar.dart';
import 'package:phone/features/groups/controllers/lesson_details_controller.dart';
import 'package:phone/features/groups/models/lesson_attendance.dart';
import 'package:phone/features/groups/pages/lesson_date_card.dart';
import 'package:phone/features/groups/pages/lesson_menu.dart';
import 'package:phone/features/groups/pages/save_lesson_action.dart';
import 'package:phone/features/groups/pages/student_attendance_card.dart';
import 'package:phone/features/groups/widgets/lesson_attendance_bottom_sheet.dart';
import 'package:phone/features/students/controllers/student_lesson_mutation_controller.dart';
import 'package:phone/features/students/controllers/student_list_controller.dart';
import 'package:phone/features/students/widgets/delete_student_lesson_bottom_sheet.dart';
import 'package:phone/generated/models/group_student_summary_response.dart';
import 'package:phone/generated/models/lesson_details_response.dart';
import 'package:phone/generated/models/lesson_response.dart';
import 'package:phone/generated/models/visit_response.dart';
import 'package:phone/generated/models/visit_status.dart';
import 'package:phone/generated/models/visit_type.dart';
import 'package:phone/i18n/strings.g.dart';

class StudentLessonPage extends ConsumerStatefulWidget {
  final String studentId;

  final String studentName;

  final DateTime date;

  final LessonResponse? lesson;

  final bool? free;

  const new({
    super.key,
    required this.studentId,
    required this.studentName,
    required this.date,
    this.lesson,
    this.free,
  });

  @override
  ConsumerState<StudentLessonPage> createState() => _StudentLessonPageState();
}

class _StudentLessonPageState extends ConsumerState<StudentLessonPage> {
  late DateTime _selectedDate;

  Map<String, LessonAttendance>? _pickedAttendance;

  late bool _isEditing = widget.lesson == null;

  @override
  void initState() {
    super.initState();
    _selectedDate = widget.lesson?.date ?? widget.date;
  }

  Map<String, LessonAttendance> _storedAttendance(
    AsyncValue<LessonDetailsResponse>? state,
  ) {
    final visits = state?.value?.studentVisits ?? const <VisitResponse>[];
    final visit = visits
        .where((visit) => visit.studentId == widget.studentId)
        .firstOrNull;

    if (visit == null) return const {};

    return {
      widget.studentId: LessonAttendance(
        status: visit.status,
        type: visit.type,
      ),
    };
  }

  bool _isStudentFree(AsyncValue<LessonDetailsResponse>? detailsState) {
    if (widget.free != null) return widget.free!;

    final students = ref.watch(studentListControllerProvider).value;
    final student = students
        ?.where((s) => s.id == widget.studentId)
        .firstOrNull;
    if (student != null) return student.free;

    final detailStudent = detailsState?.value?.students
        .where((s) => s.id == widget.studentId)
        .firstOrNull;
    if (detailStudent != null) return detailStudent.free;

    return false;
  }

  LessonAttendance _defaultAttendance(bool isFree) => LessonAttendance(
    status: VisitStatus.present,
    type: isFree ? VisitType.free : VisitType.regular,
  );

  Map<String, LessonAttendance> _attendance(
    Map<String, LessonAttendance> stored,
    bool isFree,
  ) {
    final picked = _pickedAttendance;
    if (picked != null) return picked;

    if (widget.lesson != null) return stored;

    return {widget.studentId: _defaultAttendance(isFree)};
  }

  VisitType _typeFor(
    GroupStudentSummaryResponse student,
    LessonAttendance? current,
  ) => student.free ? VisitType.free : (current?.type ?? VisitType.regular);

  bool _hasChanges(
    Map<String, LessonAttendance> attendance,
    Map<String, LessonAttendance> stored,
  ) {
    final lesson = widget.lesson;

    if (lesson == null) return true;

    final isSameDay =
        DateUtils.dateOnly(_selectedDate) == DateUtils.dateOnly(lesson.date);

    if (!isSameDay) return true;
    if (attendance.length != stored.length) return true;

    return attendance.entries.any((entry) => stored[entry.key] != entry.value);
  }

  Future<void> _pickDate() async {
    final today = DateUtils.dateOnly(DateTime.now());
    final selectedDate = DateUtils.dateOnly(_selectedDate);

    final picked = await showDatePicker(
      context: context,
      initialDate: selectedDate.isAfter(today) ? today : selectedDate,
      firstDate: DateTime(2000),
      lastDate: today,
    );

    if (picked == null || !mounted) return;

    setState(() => _selectedDate = picked);
  }

  void _toggleStudent(
    Map<String, LessonAttendance> attendance,
    GroupStudentSummaryResponse student,
  ) => setState(() {
    final next = {...attendance};
    final current = next[student.id];

    next[student.id] = LessonAttendance(
      status: current?.status == VisitStatus.present
          ? VisitStatus.absent
          : VisitStatus.present,
      type: _typeFor(student, current),
    );

    _pickedAttendance = next;
  });

  void _toggleAll(
    Map<String, LessonAttendance> attendance,
    List<GroupStudentSummaryResponse> students,
  ) => setState(() {
    final marked = students
        .where((student) => attendance.containsKey(student.id))
        .toList();
    final allPresent =
        marked.isNotEmpty &&
        marked.every(
          (student) => attendance[student.id]?.status == VisitStatus.present,
        );

    final next = {...attendance};
    for (final student in marked) {
      next[student.id] = LessonAttendance(
        status: allPresent ? VisitStatus.absent : VisitStatus.present,
        type: _typeFor(student, next[student.id]),
      );
    }

    _pickedAttendance = next;
  });

  Future<void> _editStudent(
    GroupStudentSummaryResponse student,
    Map<String, LessonAttendance> attendance,
  ) async {
    final selection = await AppBottomSheet.show<LessonAttendanceSelection>(
      context: context,
      child: LessonAttendanceBottomSheet(
        studentName: student.fullName,
        free: student.free,
        attendance: attendance[student.id],
      ),
    );

    if (!mounted || selection == null) return;

    setState(() {
      final next = {...attendance};
      final picked = selection.attendance;

      if (picked == null) {
        next.remove(student.id);
      } else {
        next[student.id] = LessonAttendance(
          status: picked.status,
          type: _typeFor(student, picked),
        );
      }

      _pickedAttendance = next;
    });
  }

  Future<void> _onSubmit(Map<String, LessonAttendance> attendance) async {
    FocusScope.of(context).unfocus();

    final lesson = widget.lesson;
    final controller = ref.read(
      studentLessonMutationControllerProvider.notifier,
    );

    final isSaved = await (lesson == null
        ? controller.createStudentLesson(
            studentId: widget.studentId,
            date: _selectedDate,
            attendance: attendance[widget.studentId],
          )
        : controller.updateStudentLesson(
            lessonId: lesson.id,
            studentId: widget.studentId,
            date: _selectedDate,
            attendance: attendance[widget.studentId],
          ));

    if (!mounted) return;

    if (isSaved) {
      Navigator.of(context).pop();
      return;
    }

    final error = ref.read(studentLessonMutationControllerProvider).error;

    AppSnackBar.show(
      context,
      context.t.error,
      error?.toString() ?? context.t.error,
      SnackBarType.error,
    );
  }

  Future<void> _onDelete() async {
    final lesson = widget.lesson;
    if (lesson == null) return;

    final isDeleted = await AppBottomSheet.show<bool>(
      context: context,
      child: DeleteStudentLessonBottomSheet(
        lessonId: lesson.id,
        studentId: widget.studentId,
      ),
    );

    if (!mounted || isDeleted != true) return;

    Navigator.of(context).pop();
  }

  String _formatDate(DateTime date) {
    final day = date.day.toString().padLeft(2, '0');
    final month = date.month.toString().padLeft(2, '0');

    return '$day.$month.${date.year}';
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final lesson = widget.lesson;

    final detailsState = lesson == null
        ? null
        : ref.watch(lessonDetailsProvider(lesson.id));
    final isLoading = ref
        .watch(studentLessonMutationControllerProvider)
        .isLoading;

    final isAttendanceLoading = detailsState?.isLoading ?? false;
    final isFree = _isStudentFree(detailsState);
    final storedAttendance = _storedAttendance(detailsState);
    final attendance = _attendance(storedAttendance, isFree);
    final hasChanges = _hasChanges(attendance, storedAttendance);
    final canEdit = _isEditing && !isLoading && !isAttendanceLoading;

    final studentItem = ref
        .watch(studentListControllerProvider)
        .value
        ?.where((s) => s.id == widget.studentId)
        .firstOrNull;

    final student = GroupStudentSummaryResponse(
      id: widget.studentId,
      fullName: widget.studentName,
      createdAt: studentItem?.createdAt ?? DateTime.now(),
      free: isFree,
      birthdate: studentItem?.birthdate,
      phone: studentItem?.phone,
      note: studentItem?.note,
      statuses: const [],
    );
    final students = isAttendanceLoading
        ? const <GroupStudentSummaryResponse>[]
        : [student];

    return Scaffold(
      appBar: AppBar(
        title: Text(
          widget.studentName.isNotEmpty
              ? context.t.lessons.individualLessonTitle(
                  name: widget.studentName,
                )
              : context.t.lessons.individualLesson,
        ),
        actions: [
          if (lesson == null || _isEditing)
            SaveLessonAction(
              isCreate: lesson == null,
              isLoading: isLoading,
              hasChanges: hasChanges,
              studentsAvailable: students.isNotEmpty,
              onPressed: () => _onSubmit(attendance),
            ),
          if (lesson != null && !_isEditing)
            LessonMenu(
              onEdit: () => setState(() => _isEditing = true),
              onDelete: _onDelete,
            ),
          const SizedBox(width: 8),
        ],
        backgroundColor: colorScheme.surface,
        surfaceTintColor: colorScheme.surface,
        shape: const RoundedRectangleBorder(
          borderRadius: AppRadius.bottomRounded,
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(22, 16, 22, 16),
        child: AbsorbPointer(
          absorbing: isLoading,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            spacing: 16,
            children: [
              LessonDateCard(
                formattedDate: _formatDate(_selectedDate),
                canEdit: canEdit,
                onTap: _pickDate,
              ),
              StudentAttendanceCard(
                students: students,
                attendance: attendance,
                canEdit: canEdit,
                isAttendanceLoading: isAttendanceLoading,
                isLoading: isLoading,
                onToggle: (s) => _toggleStudent(attendance, s),
                onLongPress: (s) => _editStudent(s, attendance),
                onToggleAll: () => _toggleAll(attendance, students),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
