import 'package:fluentui_system_icons/fluentui_system_icons.dart';
import 'package:flutter/material.dart';
import 'package:phone/core/widgets/app_bottom_sheet.dart';
import 'package:phone/features/rosters/pages/roster_list_page.dart';
import 'package:phone/features/students/pages/student_payments_page.dart';
import 'package:phone/features/students/widgets/create_student_bottom_sheet.dart';
import 'package:phone/i18n/strings.g.dart';

class StudentListPopupMenu extends StatelessWidget {
  const new({super.key});

  @override
  Widget build(BuildContext context) => PopupMenuButton<String>(
    tooltip: '',
    offset: const Offset(-8, 0),
    color: Theme.of(context).colorScheme.surface,
    itemBuilder: (menuContext) => [
      PopupMenuItem<String>(
        onTap: () => AppBottomSheet.show(
          context: context,
          child: const CreateStudentBottomSheet(),
        ),
        child: Row(
          children: [
            const Icon(FluentIcons.person_add_24_regular, size: 20),
            const SizedBox(width: 12),
            Expanded(child: Text(context.t.students.createStudent)),
          ],
        ),
      ),
      PopupMenuItem<String>(
        onTap: () => Navigator.of(
          context,
        ).push(MaterialPageRoute(builder: (context) => const RosterListPage())),
        child: Row(
          children: [
            const Icon(Icons.format_list_bulleted_outlined, size: 20),
            const SizedBox(width: 12),
            Expanded(child: Text(context.t.rosters.temporaryRosters)),
          ],
        ),
      ),
      PopupMenuItem<String>(
        onTap: () => Navigator.of(context).push(
          MaterialPageRoute(builder: (context) => const StudentPaymentsPage()),
        ),
        child: Row(
          children: [
            const Icon(FluentIcons.money_24_regular, size: 20),
            const SizedBox(width: 12),
            Expanded(child: Text(context.t.students.payments)),
          ],
        ),
      ),
    ],
  );
}
