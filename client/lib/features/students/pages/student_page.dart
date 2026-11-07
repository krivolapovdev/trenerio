import 'package:custom_refresh_indicator/custom_refresh_indicator.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:phone/core/widgets/app_bottom_sheet.dart';
import 'package:phone/core/widgets/header_bottom_spacer.dart';
import 'package:phone/core/widgets/header_snap_scroll_physics.dart';
import 'package:phone/core/widgets/recent_transactions.dart';
import 'package:phone/features/finance/widgets/create_transaction_bottom_sheet.dart';
import 'package:phone/features/students/controllers/student_lessons_controller.dart';
import 'package:phone/features/students/controllers/student_list_controller.dart';
import 'package:phone/features/students/controllers/student_payment_list_controller.dart';
import 'package:phone/features/students/pages/student_app_menu.dart';
import 'package:phone/features/students/pages/student_group_page.dart';
import 'package:phone/features/students/pages/student_lesson_page.dart';
import 'package:phone/features/students/pages/student_payment_list_page.dart';
import 'package:phone/features/students/widgets/delete_student_bottom_sheet.dart';
import 'package:phone/features/students/widgets/edit_student_bottom_sheet.dart';
import 'package:phone/features/students/widgets/student_groups_section.dart';
import 'package:phone/features/students/widgets/student_header.dart';
import 'package:phone/features/students/widgets/student_lessons_section.dart';
import 'package:phone/features/students/widgets/student_report_bottom_sheet.dart';
import 'package:phone/generated/models/student_summary_response.dart';
import 'package:phone/generated/models/transaction_response.dart';

class StudentPage extends ConsumerStatefulWidget {
  final StudentSummaryResponse student;

  const new({super.key, required this.student});

  @override
  ConsumerState<StudentPage> createState() => _StudentPageState();
}

class _StudentPageState extends ConsumerState<StudentPage> {
  late StudentSummaryResponse _student = widget.student;
  late final ScrollController _scrollController = ScrollController();
  bool _isSnapping = false;
  bool _isRefreshing = false;

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  Future<void> _openEditStudentSheet([StudentSummaryResponse? student]) async {
    final s = student ?? _student;
    await AppBottomSheet.show(
      context: context,
      child: EditStudentBottomSheet(student: s),
    );
  }

  Future<void> _openAssignGroupPage([StudentSummaryResponse? student]) async {
    final s = student ?? _student;
    final groupIds = (s.studentGroups ?? const [])
        .map((group) => group.id)
        .toSet();
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) =>
            StudentGroupPage(studentId: s.id, initialGroupIds: groupIds),
      ),
    );
  }

  Future<void> _openCreateLessonPage([StudentSummaryResponse? student]) async {
    final s = student ?? _student;
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => StudentLessonPage(
          studentId: s.id,
          studentName: s.fullName,
          free: s.free,
          date: DateTime.now(),
        ),
      ),
    );
  }

  Future<void> _openReportSheet([StudentSummaryResponse? student]) async {
    final s = student ?? _student;
    await AppBottomSheet.show(
      context: context,
      child: StudentReportBottomSheet(student: s),
    );
  }

  Future<void> _openCreatePaymentSheet([
    StudentSummaryResponse? student,
  ]) async {
    final s = student ?? _student;
    await AppBottomSheet.show(
      context: context,
      child: CreateTransactionBottomSheet(
        initialStudentId: s.id,
        initialStudentName: s.fullName,
        initialAmount: s.studentGroups?.firstOrNull?.defaultPrice,
        isIncomeOnly: true,
        isStudentLocked: true,
      ),
    );
  }

  Future<void> _confirmDelete([StudentSummaryResponse? student]) async {
    final s = student ?? _student;
    final navigator = Navigator.of(context);

    final deleted = await AppBottomSheet.show<bool>(
      context: context,
      child: DeleteStudentBottomSheet(student: s),
    );

    if (!mounted || deleted != true) return;

    navigator.pop();
  }

  Future<void> _refresh([String? studentId]) async {
    final id = studentId ?? _student.id;
    setState(() => _isRefreshing = true);

    try {
      await Future.wait([
        ref.read(studentLessonsProvider(id).notifier).refresh(),
        ref.read(studentPaymentsControllerProvider(id).notifier).refresh(),
        ref
            .read(studentListControllerProvider.notifier)
            .getAllStudents(forceRefresh: true),
      ]);

      if (!mounted) return;

      final listState = ref.read(studentListControllerProvider);
      if (listState.hasError) return;

      final updated = (listState.value ?? const [])
          .where((student) => student.id == id)
          .firstOrNull;

      if (updated == null) {
        Navigator.of(context).pop();
        return;
      }

      setState(() => _student = updated);
    } finally {
      if (mounted) {
        setState(() => _isRefreshing = false);
      }
    }
  }

  bool _handleScrollNotification(ScrollNotification notification) {
    if (notification is ScrollEndNotification) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _snapIfNeeded();
      });
    }
    return false;
  }

  void _snapIfNeeded() {
    if (!mounted || !_scrollController.hasClients || _isSnapping) return;
    final position = _scrollController.position;

    // Do not snap if overscrolling (pull-to-refresh)
    if (position.outOfRange) return;

    final offset = position.pixels;
    const snapOffset = StudentHeaderDelegate.contentHeight;

    if (offset > 1.0 && offset < snapOffset - 1.0) {
      final target = offset < (snapOffset / 2) ? 0.0 : snapOffset;
      _isSnapping = true;
      _scrollController
          .animateTo(
            target,
            duration: const Duration(milliseconds: 200),
            curve: Curves.easeOutCubic,
          )
          .whenComplete(() {
            _isSnapping = false;
          });
    }
  }

  @override
  Widget build(BuildContext context) {
    final paymentsAsync = ref.watch(
      studentPaymentsControllerProvider(_student.id),
    );

    final studentFromList = ref.watch(
      studentListControllerProvider.select(
        (state) => (state.value ?? const <StudentSummaryResponse>[])
            .where((student) => student.id == _student.id)
            .firstOrNull,
      ),
    );
    final currentStudent = studentFromList ?? _student;

    ref.listen(studentListControllerProvider, (previous, next) {
      final updated = (next.value ?? const <StudentSummaryResponse>[])
          .where((student) => student.id == _student.id)
          .firstOrNull;

      if (updated != null && mounted) {
        setState(() => _student = updated);
      }
    });

    final topPadding = MediaQuery.paddingOf(context).top;
    final headerMaxHeight =
        topPadding + kToolbarHeight + StudentHeaderDelegate.contentHeight;

    const overlayStyle = SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      statusBarBrightness: Brightness.light,
    );

    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: overlayStyle,
      child: Scaffold(
        body: CustomMaterialIndicator(
          edgeOffset: headerMaxHeight,
          displacement: 20,
          color: Colors.black,
          clipBehavior: Clip.antiAlias,
          onRefresh: () => _refresh(currentStudent.id),
          child: NotificationListener<ScrollNotification>(
            onNotification: _handleScrollNotification,
            child: CustomScrollView(
              controller: _scrollController,
              clipBehavior: Clip.none,
              physics: const HeaderSnapScrollPhysics(
                snapOffset: StudentHeaderDelegate.contentHeight,
                parent: BouncingScrollPhysics(
                  parent: AlwaysScrollableScrollPhysics(),
                ),
              ),
              slivers: [
                SliverPersistentHeader(
                  pinned: true,
                  delegate: StudentHeaderDelegate(
                    student: currentStudent,
                    payments:
                        paymentsAsync.value ?? const <TransactionResponse>[],
                    topPadding: topPadding,
                    isRefreshing: _isRefreshing,
                    menu: StudentAppMenu(
                      onEdit: () => _openEditStudentSheet(currentStudent),
                      onCreateLesson: () =>
                          _openCreateLessonPage(currentStudent),
                      onCreatePayment: () =>
                          _openCreatePaymentSheet(currentStudent),
                      onAssignGroup: () => _openAssignGroupPage(currentStudent),
                      onReport: () => _openReportSheet(currentStudent),
                      onDelete: () => _confirmDelete(currentStudent),
                    ),
                  ),
                ),
                SliverPadding(
                  padding: const EdgeInsets.all(16),
                  sliver: SliverList.list(
                    children: [
                      StudentLessonsSection(
                        studentId: currentStudent.id,
                        studentName: currentStudent.fullName,
                        studentGroups: currentStudent.studentGroups ?? const [],
                      ),
                      const SizedBox(height: 16),
                      RecentTransactions(
                        asyncTransactions: paymentsAsync,
                        overrideTitle: currentStudent.fullName,
                        onSeeAllPressed: () {
                          Navigator.of(context).push(
                            MaterialPageRoute(
                              builder: (context) => StudentPaymentListPage(
                                student: currentStudent,
                              ),
                            ),
                          );
                        },
                        onAddPressed: () =>
                            _openCreatePaymentSheet(currentStudent),
                      ),
                      const SizedBox(height: 16),
                      StudentGroupsSection(
                        student: currentStudent,
                        onEdit: () => _openAssignGroupPage(currentStudent),
                      ),
                    ],
                  ),
                ),
                const HeaderBottomSpacer(
                  snapOffset: StudentHeaderDelegate.contentHeight,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
